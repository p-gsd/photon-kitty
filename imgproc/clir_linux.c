// +build linux

#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <time.h>

const char *source = "\n\
__constant sampler_t sampler = CLK_NORMALIZED_COORDS_FALSE | CLK_ADDRESS_CLAMP_TO_EDGE | CLK_FILTER_NEAREST;\n\
inline float4 sample_bicubic_border(__read_only image2d_t source, float2 pos, int2 SrcSize)\n\
{\n\
   int2 isrcpos = convert_int2(pos);\n\
   float dx = pos.x - isrcpos.x;\n\
   float dy = pos.y - isrcpos.y;\n\
\n\
   float4 C[4] = {0, 0, 0, 0};\n\
\n\
   if (isrcpos.x < 0 || isrcpos.x >= SrcSize.x)\n\
      return 0;\n\
\n\
   if (isrcpos.y < 0 || isrcpos.y >= SrcSize.y)\n\
      return 0;\n\
\n\
   for (int i = 0; i < 4; i++)\n\
   {\n\
      int y = isrcpos.y - 1 + i;\n\
      if (y < 0)\n\
         y = 0;\n\
\n\
      if (y >= SrcSize.y)\n\
         y = SrcSize.y - 1;\n\
\n\
      int Middle = clamp(isrcpos.x, 0, SrcSize.x - 1);\n\
\n\
      const int2 pos0 = { Middle, y };\n\
      float4 center = read_imagef(source, sampler, pos0);\n\
\n\
      float4 left = 0, right1 = 0, right2 = 0;\n\
      if (isrcpos.x - 1 >= 0)\n\
      {\n\
         const int2 pos1 = { isrcpos.x - 1, y };\n\
         left = read_imagef(source, sampler, pos1);\n\
      }\n\
      else\n\
      {\n\
         left = center;\n\
      }\n\
\n\
      if (isrcpos.x + 1 < SrcSize.x)\n\
      {\n\
         const int2 pos2 = { isrcpos.x + 1, y };\n\
         right1 = read_imagef(source, sampler, pos2);\n\
      }\n\
      else\n\
      {\n\
         right1 = center;\n\
      }\n\
\n\
      if (isrcpos.x + 2 < SrcSize.x)\n\
      {\n\
         const int2 pos3 = { isrcpos.x + 2, y };\n\
         right2 = read_imagef(source, sampler, pos3);\n\
      }\n\
      else\n\
      {\n\
         right2 = right1;\n\
      }\n\
\n\
      float4 a0 = center;\n\
      float4 d0 = left - a0;\n\
      float4 d2 = right1 - a0;\n\
      float4 d3 = right2 - a0;\n\
\n\
      float4 a1 = -1.0f / 3 * d0 + d2 - 1.0f / 6 * d3;\n\
      float4 a2 =  1.0f / 2 * d0 + 1.0f / 2 * d2;\n\
      float4 a3 = -1.0f / 6 * d0 - 1.0f / 2 * d2 + 1.0f / 6 * d3;\n\
      C[i] = a0 + a1 * dx + a2 * dx * dx + a3 * dx * dx * dx;\n\
   }\n\
\n\
   float4 d0 = C[0] - C[1];\n\
   float4 d2 = C[2] - C[1];\n\
   float4 d3 = C[3] - C[1];\n\
   float4 a0 = C[1];\n\
   float4 a1 = -1.0f / 3 * d0 + d2 -1.0f / 6 * d3;\n\
   float4 a2 = 1.0f / 2 * d0 + 1.0f / 2 * d2;\n\
   float4 a3 = -1.0f / 6 * d0 - 1.0f / 2 * d2 + 1.0f / 6 * d3;\n\
\n\
   return a0 + a1 * dy + a2 * dy * dy + a3 * dy * dy * dy;\n\
}\n\
\n\
__kernel void resize_bicubic(read_only image2d_t source, write_only image2d_t dest, uint src_img_width, uint src_img_height, uint dst_img_width, uint dst_img_height, float ratioX, float ratioY)\n\
{\n\
   const int gx = get_global_id(0);\n\
   const int gy = get_global_id(1);\n\
   const int2 pos = { gx, gy };\n\
\n\
   if (pos.x >= dst_img_width || pos.y >= dst_img_height)\n\
      return;\n\
\n\
   float2 srcpos = {(pos.x + 0.4995f) / ratioX, (pos.y + 0.4995f) / ratioY};\n\
   int2 SrcSize = { (int)src_img_width, (int)src_img_height };\n\
\n\
   float4 value;\n\
\n\
   srcpos -= (float2)(0.5f, 0.5f);\n\
\n\
   int2 isrcpos = convert_int2(srcpos);\n\
   float dx = srcpos.x - isrcpos.x;\n\
   float dy = srcpos.y - isrcpos.y;\n\
\n\
   if (isrcpos.x <= 0 || isrcpos.x >= SrcSize.x - 2)\n\
      value = sample_bicubic_border(source, srcpos, SrcSize);\n\
\n\
   if (isrcpos.y <= 0 || isrcpos.y >= SrcSize.y - 2)\n\
      value = sample_bicubic_border(source, srcpos, SrcSize);\n\
\n\
   float4 C[4] = {0, 0, 0, 0};\n\
\n\
   for (int i = 0; i < 4; i++)\n\
   {\n\
      const int y = isrcpos.y - 1 + i;\n\
\n\
      const int2 pos0 = { isrcpos.x, y };\n\
      const int2 pos1 = { isrcpos.x - 1, y };\n\
      const int2 pos2 = { isrcpos.x + 1, y };\n\
      const int2 pos3 = { isrcpos.x + 2, y };\n\
\n\
      float4 a0 = read_imagef(source, sampler, pos0);\n\
      float4 d0 = read_imagef(source, sampler, pos1) - a0;\n\
      float4 d2 = read_imagef(source, sampler, pos2) - a0;\n\
      float4 d3 = read_imagef(source, sampler, pos3) - a0;\n\
\n\
      float4 a1 = -1.0f / 3 * d0 + d2 - 1.0f / 6 * d3;\n\
      float4 a2 =  1.0f / 2 * d0 + 1.0f / 2 * d2;\n\
      float4 a3 = -1.0f / 6 * d0 - 1.0f / 2 * d2 + 1.0f / 6 * d3;\n\
      C[i] = a0 + a1 * dx + a2 * dx * dx + a3 * dx * dx * dx;\n\
   }\n\
\n\
   float4 d0 = C[0] - C[1];\n\
   float4 d2 = C[2] - C[1];\n\
   float4 d3 = C[3] - C[1];\n\
   float4 a0 = C[1];\n\
   float4 a1 = -1.0f / 3 * d0 + d2 -1.0f / 6 * d3;\n\
   float4 a2 = 1.0f / 2 * d0 + 1.0f / 2 * d2;\n\
   float4 a3 = -1.0f / 6 * d0 - 1.0f / 2 * d2 + 1.0f / 6 * d3;\n\
   value = a0 + a1 * dy + a2 * dy * dy + a3 * dy * dy * dy;\n\
\n\
   write_imagef(dest, pos, value);\n\
}\n\
\n\
__kernel void assign_to_centers(\n\
		read_only image2d_t img,\n\
		uint img_width,\n\
		__global float *centers,\n\
		uint k,\n\
		__global uchar *assignments,\n\
		__global int *changed\n\
) {\n\
   const int gx = get_global_id(0);\n\
   const int gy = get_global_id(1);\n\
   const int2 pos = { gx, gy };\n\
   const float4 value = read_imagef(img, sampler, pos);\n\
   const int ai = (img_width*gy)+gx;\n\
   float dd = 1.0E30f;\n\
   uchar indMin = -1;\n\
   for (int i = 0; i < k; i++) {\n\
	   float4 c = {centers[i*4],centers[i*4+1],centers[i*4+2],centers[i*4+3]};\n\
       float localD = length(c - value);\n\
	   if (localD < dd) {\n\
		   indMin = i;\n\
		   dd = localD;\n\
	   }\n\
   }\n\
   if (assignments[ai] != indMin) changed[0] = 1;\n\
   assignments[ai] = indMin;\n\
}\n\
\n\
__kernel void recompute_centers(\n\
		read_only image2d_t img,\n\
		uint img_width,\n\
		uint img_height,\n\
		__global float *centers,\n\
		__global ulong *global_sum,\n\
		__global uint *global_count,\n\
		__global uchar *assignments,\n\
		__local float4 *sum,\n\
		__local uint *count\n\
) {\n\
   const int window = 8;\n\
   const int gid = get_global_id(0);\n\
   const int lid = get_local_id(0);\n\
   const int local_size = get_local_size(0);\n\
   sum[lid] = 0;\n\
   count[lid] = 0;\n\
   global_sum[lid] = 0;\n\
   global_count[lid] = 0;\n\
   const int item_offset = gid*window;\n\
   for (int x = item_offset; x < (item_offset + window) && x < img_width; x++) {\n\
      for (int y = item_offset; y < (item_offset + window) && y < img_height; y++) {\n\
		 if (lid == assignments[img_width*y+x]) {\n\
			 const int2 pos = {x,y};\n\
			 const float4 value = read_imagef(img, sampler, pos);\n\
			 sum[lid] += value;\n\
			 count[lid]++;\n\
		 }\n\
	  }\n\
   }\n\
   barrier(CLK_GLOBAL_MEM_FENCE);\n\
   if (sum[lid].x != 0) atom_add(global_sum + lid*4, (int)(sum[lid].x*255));\n\
   if (sum[lid].y != 0) atom_add(global_sum + lid*4+1, (int)(sum[lid].y*255));\n\
   if (sum[lid].z != 0) atom_add(global_sum + lid*4+2, (int)(sum[lid].z*255));\n\
   if (sum[lid].w != 0) atom_add(global_sum + lid*4+3, (int)(sum[lid].w*255));\n\
   if (count[lid] != 0) atomic_add(global_count + lid, count[lid]);\n\
   if (gid == 0) {\n\
	   for (int i=0; i<local_size; i++) {\n\
		   if (global_count[i] == 0) continue;\n\
	       centers[i*4] = global_sum[i*4] / global_count[i] / 255.0;\n\
	       centers[i*4+1] = global_sum[i*4+1] / global_count[i] / 255.0;\n\
	       centers[i*4+2] = global_sum[i*4+2] / global_count[i] / 255.0;\n\
	       centers[i*4+3] = global_sum[i*4+3] / global_count[i] / 255.0;\n\
	   }\n\
   }\n\
}\n\
\n\
//pseudo random generator; seed passed to kernel + globalID\n\
inline uint rand(uint seed) {\n\
	seed = (seed * 0x5DEECE66DL + 0xBL) & ((1L << 48) - 1);\n\
	uint result = seed >> 16;\n\
	return result;\n\
}\n\
\n\
__kernel void random_centers_from_image(\n\
		read_only image2d_t img,\n\
		uint img_width,\n\
		uint img_height,\n\
		__global float *centers,\n\
		uint seed\n\
	) {\n\
   const int i = get_global_id(0);\n\
   const uint x = rand(seed+i*2) \% img_width;\n\
   const uint y = rand(seed+i*3) \% img_height;\n\
   const int2 pos = { x, y };\n\
   const float4 value = read_imagef(img, sampler, pos);\n\
   centers[i*4] = value.x;\n\
   centers[i*4+1] = value.y;\n\
   centers[i*4+2] = value.z;\n\
   centers[i*4+3] = value.w;\n\
}\n\
\n\
__kernel void convert_to_uint8(\n\
		__global float *centers,\n\
		__global uchar *palette\n\
	) {\n\
   const int i = get_global_id(0);\n\
   palette[i*4] = centers[i*4] * 255;\n\
   palette[i*4+1] = centers[i*4+1] * 255;\n\
   palette[i*4+2] = centers[i*4+2] * 255;\n\
   palette[i*4+3] = centers[i*4+3] * 255;\n\
}\n\
\n\
__kernel void add_one(__global uchar *assignments) {\n\
   assignments[get_global_id(0)] += 1;\n\
}";

// openCL lib pointer
void *libOpenCL;

// openCL constants
#define CL_INTENSITY 0x10B8
#define CL_RGBA 0x10B5
#define CL_UNORM_INT8 0x10D2
#define CL_MEM_READ_WRITE (1 << 0)
#define CL_MEM_WRITE_ONLY (1 << 1)
#define CL_MEM_READ_ONLY (1 << 2)
#define CL_MEM_COPY_HOST_PTR (1 << 5)
#define CL_MEM_OBJECT_IMAGE2D 0x10F1
#define CL_DEVICE_TYPE_GPU (1 << 2)
#define CL_PROGRAM_BUILD_LOG 0x1183
#define CL_DEVICE_IMAGE_SUPPORT 0x1016

// openCL types
typedef uint8_t cl_uchar;
typedef int32_t cl_int;
typedef uint32_t cl_uint;
typedef uint64_t cl_ulong;
typedef cl_uint cl_bool;
typedef cl_ulong cl_bitfield;
typedef cl_bitfield cl_device_type;
typedef intptr_t cl_context_properties;
typedef cl_bitfield cl_mem_flags;
typedef void *cl_mem;
typedef struct _cl_image_format {
	cl_uint image_channel_order;
	cl_uint image_channel_data_type;
} cl_image_format;
typedef struct _cl_image_desc {
	cl_uint image_type;
	size_t image_width;
	size_t image_height;
	size_t image_depth;
	size_t image_array_size;
	size_t image_row_pitch;
	size_t image_slice_pitch;
	cl_uint num_mip_levels;
	cl_uint num_samples;
	void *buffer;
} cl_image_desc;

// openCL function pointers
cl_int(*clGetPlatformIDs) (cl_uint num_entries, void **platforms, cl_uint * num_platforms);
cl_int(*clGetDeviceIDs) (void *platform, cl_device_type device_type, cl_uint num_entries, void **devices,
			 cl_uint * num_devices);
void *(*clCreateContext)(cl_context_properties * properties, cl_uint num_devices, void **devices, void *callback,
			 void *user_data, cl_int * errcode_ret);
void *(*clCreateProgramWithSource)(void *context, cl_uint count, const char **strings, size_t *lengths,
				   cl_int * errcode_ret);
cl_mem(*clCreateImage) (void *context, cl_mem_flags flags, cl_image_format * image_format, cl_image_desc * image_desc,
			void *host_ptr, cl_int * errcode_ret);
void *(*clCreateCommandQueueWithProperties)(void *context, void *device, void *properties, cl_int * errcode_ret);
cl_int(*clBuildProgram) (void *program, cl_uint num_devices, void **device_list, char *options, void *,
			 void *user_data);
cl_int(*clGetProgramBuildInfo) (void *program, void *device, cl_uint param_name, size_t param_value_size,
				void *param_value, size_t *param_value_size_ret);
cl_int(*clReleaseMemObject) (cl_mem memobj);
cl_int(*clReleaseKernel) (void *kernel);
void *(*clCreateKernel)(void *program, const char *kernel_name, cl_int * errcode_ret);
cl_int(*clSetKernelArg) (void *kernel, cl_uint arg_index, size_t arg_size, const void *arg_value);
cl_int(*clEnqueueNDRangeKernel) (void *command_queue, void *kernel, cl_uint work_dim, const size_t *global_work_offset,
				 const size_t *global_work_size, const size_t *local_work_size,
				 cl_uint num_events_in_wait_list, const void **event_wait_list, void **event);
cl_int(*clEnqueueReadImage) (void *command_queue, cl_mem image, cl_bool blocking_read, const size_t *origin,
			     const size_t *region, size_t row_pitch, size_t slice_pitch, void *ptr,
			     cl_uint num_events_in_wait_list, void **event_wait_list, void **event);
cl_int(*clWaitForEvents) (cl_uint num_events, void **event_list);
cl_int(*clReleaseEvent) (void *event);
cl_mem(*clCreateBuffer) (void *context, cl_mem_flags flags, size_t size, void *host_ptr, cl_int * errcode_ret);
cl_int(*clEnqueueReadBuffer) (void *command_queue,
			      cl_mem buffer,
			      cl_bool blocking_read,
			      size_t offset,
			      size_t size,
			      void *ptr, cl_uint num_events_in_wait_list, void *event_wait_list, void *event);
cl_int(*clEnqueueWriteBuffer) (void *command_queue,
			       cl_mem buffer,
			       cl_bool blocking_write,
			       size_t offset,
			       size_t size,
			       const void *ptr, cl_uint num_events_in_wait_list, void *event_wait_list, void *event);
cl_int(*clGetDeviceInfo) (void *device,
			  cl_uint param_name, size_t param_value_size, void *param_value, size_t *param_value_size_ret);

// global variables for the device and program
void *deviceID;
void *deviceCtx;
void *queue;
void *program;

// load libOpenCL.so, load functions, initialize device and program
int init()
{
	if (libOpenCL) {
		return 0;
	}

	libOpenCL = dlopen("libOpenCL.so", RTLD_LAZY);

	if (!libOpenCL) {
		return -1;
	}

	clGetPlatformIDs = (cl_int(*)())dlsym(libOpenCL, "clGetPlatformIDs");
	if (clGetPlatformIDs == NULL) {
		return -1;
	}
	clGetDeviceIDs = (cl_int(*)())dlsym(libOpenCL, "clGetDeviceIDs");
	if (clGetDeviceIDs == NULL) {
		return -1;
	}
	clGetDeviceInfo = (cl_int(*)(void *, cl_uint, size_t, void *, size_t *))dlsym(libOpenCL, "clGetDeviceInfo");
	if (clGetDeviceInfo == NULL) {
		return -1;
	}
	clCreateContext = (void *(*)())dlsym(libOpenCL, "clCreateContext");
	if (clCreateContext == NULL) {
		return -1;
	}
	clCreateProgramWithSource = (void *(*)())dlsym(libOpenCL, "clCreateProgramWithSource");
	clCreateImage = (void *(*)())dlsym(libOpenCL, "clCreateImage");
	if (clCreateImage == NULL) {
		return -1;
	}
	clCreateCommandQueueWithProperties = (void *(*)())dlsym(libOpenCL, "clCreateCommandQueueWithProperties");
	if (clCreateCommandQueueWithProperties == NULL) {
		return -1;
	}
	clBuildProgram =
	    (cl_int(*)(void *, cl_uint, void **, char *, void *, void *))dlsym(libOpenCL, "clBuildProgram");
	if (clBuildProgram == NULL) {
		return -1;
	}
	clGetProgramBuildInfo =
	    (cl_int(*)(void *, void *, cl_uint, size_t, void *, size_t *))dlsym(libOpenCL, "clGetProgramBuildInfo");
	if (clGetProgramBuildInfo == NULL) {
		return -1;
	}
	clReleaseMemObject = (cl_int(*)(void *))dlsym(libOpenCL, "clReleaseMemObject");
	if (clReleaseMemObject == NULL) {
		return -1;
	}
	clReleaseKernel = (cl_int(*)(void *))dlsym(libOpenCL, "clReleaseKernel");
	if (clReleaseKernel == NULL) {
		return -1;
	}
	clCreateKernel = (void *(*)(void *, const char *, cl_int *))dlsym(libOpenCL, "clCreateKernel");
	if (clCreateKernel == NULL) {
		return -1;
	}
	clSetKernelArg = (cl_int(*)(void *, cl_uint, size_t, const void *))dlsym(libOpenCL, "clSetKernelArg");
	if (clSetKernelArg == NULL) {
		return -1;
	}
	clEnqueueNDRangeKernel = (cl_int(*)
				  (void *, void *, cl_uint, const size_t *, const size_t *,
				   const size_t *, cl_uint, const void **, void **))dlsym(libOpenCL,
											  "clEnqueueNDRangeKernel");
	if (clEnqueueNDRangeKernel == NULL) {
		return -1;
	}
	clEnqueueReadImage = (cl_int(*)
			      (void *, cl_mem, cl_bool, const size_t *, const size_t *, size_t,
			       size_t, void *, cl_uint, void **, void **))dlsym(libOpenCL, "clEnqueueReadImage");
	if (clEnqueueReadImage == NULL) {
		return -1;
	}
	clWaitForEvents = (cl_int(*)(cl_uint, void **))dlsym(libOpenCL, "clWaitForEvents");
	if (clWaitForEvents == NULL) {
		return -1;
	}
	clReleaseEvent = (cl_int(*)(void *))dlsym(libOpenCL, "clReleaseEvent");
	if (clReleaseEvent == NULL) {
		return -1;
	}
	clCreateBuffer = (cl_mem(*)(void *, cl_mem_flags, size_t, void *, cl_int *))dlsym(libOpenCL, "clCreateBuffer");
	if (clReleaseEvent == NULL) {
		return -1;
	}
	clEnqueueReadBuffer =
	    (cl_int(*)(void *, void *, cl_bool, size_t, size_t, void *, cl_uint, void *, void *))dlsym(libOpenCL,
												       "clEnqueueReadBuffer");
	if (clEnqueueReadBuffer == NULL) {
		return -1;
	}
	clEnqueueWriteBuffer =
	    (cl_int(*)(void *, void *, cl_bool, size_t, size_t, const void *, cl_uint, void *, void *))dlsym(libOpenCL,
													     "clEnqueueWriteBuffer");
	if (clEnqueueWriteBuffer == NULL) {
		return -1;
	}

	int ret;
	unsigned int len;
	ret = (*clGetPlatformIDs) (0, NULL, &len);
	if (ret != 0)
		return ret;

	void **platformIDs = malloc(len * sizeof(void *));
	ret = (*clGetPlatformIDs) (len, platformIDs, NULL);
	if (ret != 0)
		return ret;
	for (int i = 0; i < len; i++) {
		ret = (*clGetDeviceIDs) (platformIDs[i], CL_DEVICE_TYPE_GPU, 1, &deviceID, NULL);
		if (ret == 0) {
			cl_uint image_support;
			(*clGetDeviceInfo) (deviceID, CL_DEVICE_IMAGE_SUPPORT, sizeof(image_support), &image_support,
					    NULL);
			if (image_support == 1)
				break;
		}
		deviceID = NULL;
	}
	free(platformIDs);

	if (deviceID == NULL) {
		return -100;
	}

	deviceCtx = (*clCreateContext) (NULL, 1, &deviceID, NULL, NULL, &ret);
	if (ret != 0)
		return ret;
	queue = (*clCreateCommandQueueWithProperties) (deviceCtx, deviceID, NULL, &ret);
	if (ret != 0)
		return ret;
	program = (*clCreateProgramWithSource) (deviceCtx, 1, &source, NULL, &ret);
	if (ret != 0)
		return ret;
	ret = (*clBuildProgram) (program, 1, &deviceID, NULL, NULL, NULL);
	if (ret != 0) {
		if (ret == -11) {
			size_t len;
			cl_int r = (*clGetProgramBuildInfo) (program, deviceID, CL_PROGRAM_BUILD_LOG, 0, NULL, &len);
			if (r != 0)
				return ret;
			char log[2048];
			(*clGetProgramBuildInfo) (program, deviceID, CL_PROGRAM_BUILD_LOG, sizeof(log), log, NULL);
			printf("\n%s\n", log);
		}
		return ret;
	}
	time_t t;
	srand((unsigned)time(&t));

	return 0;
}

size_t max(int a, int b)
{
	return (a > b) ? a : b;
}

typedef struct {
	void *img;
	unsigned int width;
	unsigned int height;
	unsigned int rowPitch;
	unsigned int pixelSize;
	void *resizeKernel;
	void *assignToCentersKernel;
	void *recomputeCentersKernel;
	void *randomCentersFromImageKernel;
	void *convertToUint8Kernel;
	void *addOneKernel;
} imageResizer;

int releaseImageResizer(void *irp)
{
	imageResizer *ir = (imageResizer *) irp;

	if (ir->img != NULL) {
		cl_int ret = (*clReleaseMemObject) (ir->img);
		if (ret != 0)
			return ret;
	}
	if (ir->resizeKernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->resizeKernel);
		if (ret != 0)
			return ret;
	}
	if (ir->assignToCentersKernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->assignToCentersKernel);
		if (ret != 0)
			return ret;
	}
	if (ir->recomputeCentersKernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->recomputeCentersKernel);
		if (ret != 0)
			return ret;
	}
	if (ir->randomCentersFromImageKernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->randomCentersFromImageKernel);
		if (ret != 0)
			return ret;
	}
	if (ir->convertToUint8Kernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->convertToUint8Kernel);
		if (ret != 0)
			return ret;
	}
	if (ir->addOneKernel != NULL) {
		cl_int ret = (*clReleaseKernel) (ir->addOneKernel);
		if (ret != 0)
			return ret;
	}
	free(ir);
	return 0;
}

void *createImageResizer(unsigned int width, unsigned int height,
			 unsigned int rowPitch, unsigned int pixelSize, void *data, cl_int * ret)
{
	cl_image_format format;
	format.image_channel_order = pixelSize == 1 ? CL_INTENSITY : CL_RGBA;
	format.image_channel_data_type = CL_UNORM_INT8;
	cl_mem_flags inFlags = CL_MEM_READ_ONLY | CL_MEM_COPY_HOST_PTR;
	cl_image_desc inDesc;
	inDesc.image_type = CL_MEM_OBJECT_IMAGE2D;
	inDesc.image_width = width;
	inDesc.image_height = height;
	inDesc.image_depth = 0;
	inDesc.image_array_size = 0;
	inDesc.image_row_pitch = (size_t)(rowPitch);
	inDesc.image_slice_pitch = 0;
	inDesc.num_mip_levels = 0;
	inDesc.num_samples = 0;
	inDesc.buffer = NULL;
	imageResizer *ir = calloc(1, sizeof(imageResizer));
	ir->img = (*clCreateImage) (deviceCtx, inFlags, &format, &inDesc, data, ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->width = width;
	ir->height = height;
	ir->rowPitch = rowPitch;

	ir->resizeKernel = (*clCreateKernel) (program, "resize_bicubic", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->assignToCentersKernel = (*clCreateKernel) (program, "assign_to_centers", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->recomputeCentersKernel = (*clCreateKernel) (program, "recompute_centers", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->randomCentersFromImageKernel = (*clCreateKernel) (program, "random_centers_from_image", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->convertToUint8Kernel = (*clCreateKernel) (program, "convert_to_uint8", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}
	ir->addOneKernel = (*clCreateKernel) (program, "add_one", ret);
	if (*ret != 0) {
		releaseImageResizer((void *)ir);
		return ret;
	}

	return (void *)(ir);
}

int _resize(imageResizer * ir, unsigned int outWidth, unsigned int outHeight, cl_mem * out, void **event)
{
	cl_int ret;
	cl_image_format format;
	format.image_channel_order = ir->pixelSize == 1 ? CL_INTENSITY : CL_RGBA;
	format.image_channel_data_type = CL_UNORM_INT8;
	// output image
	cl_mem_flags outFlags = CL_MEM_WRITE_ONLY;
	cl_image_desc outDesc;
	outDesc.image_type = CL_MEM_OBJECT_IMAGE2D;
	outDesc.image_width = outWidth;
	outDesc.image_height = outHeight;
	outDesc.image_row_pitch = 0;	// rowPitch;
	outDesc.image_slice_pitch = 0;
	outDesc.num_mip_levels = 0;
	outDesc.num_samples = 0;
	outDesc.buffer = NULL;
	*out = (*clCreateImage) (deviceCtx, outFlags, &format, &outDesc, NULL, &ret);
	if (ret != 0)
		return ret;
	// set args to resize kernel
	ret = (*clSetKernelArg) (ir->resizeKernel, 0, sizeof(void *), &ir->img);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 1, sizeof(void *), out);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 2, sizeof(unsigned int), &ir->width);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 3, sizeof(unsigned int), &ir->height);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 4, sizeof(unsigned int), &outWidth);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 5, sizeof(unsigned int), &outHeight);
	float ratioX = (float)(outWidth) / (float)(ir->width);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 6, sizeof(float), &ratioX);
	float ratioY = (float)(outHeight) / (float)(ir->height);
	ret |= (*clSetKernelArg) (ir->resizeKernel, 7, sizeof(float), &ratioY);
	if (ret != 0)
		return ret;
	// run kernel
	size_t globalSize[] = { max(ir->width, outWidth), max(ir->height, outHeight) };
	ret = (*clEnqueueNDRangeKernel) (queue, ir->resizeKernel, 2, 0, globalSize, 0, 0, NULL, event);
	if (ret != 0)
		return ret;
}

int resize(void *irp, unsigned int outWidth, unsigned int outHeight, void *outData)
{
	imageResizer *ir = (imageResizer *) irp;
	cl_mem out;
	void *kernelEvent;
	cl_int ret = _resize(ir, outWidth, outHeight, &out, &kernelEvent);
	if (ret != 0)
		return ret;
	// get image
	void *readEvent;
	size_t origin[] = { 0, 0, 0 };
	size_t region[] = { outWidth, outHeight, 1 };
	ret = (*clEnqueueReadImage) (queue, out, 1, origin, region, 0, 0, outData, 1, &kernelEvent, &readEvent);
	ret |= (*clWaitForEvents) (1, &readEvent);
	ret |= (*clReleaseEvent) (kernelEvent);
	ret |= (*clReleaseEvent) (readEvent);
	ret |= (*clReleaseMemObject) (out);
	if (ret != 0)
		return ret;

	return 0;
}

int getImageData(cl_mem buf, size_t w, size_t h, void *ptr)
{
	cl_int ret;
	void *readEvent;
	size_t origin[] = { 0, 0, 0 };
	size_t region[] = { w, h, 1 };
	ret = (*clEnqueueReadImage) (queue, buf, 1, origin, region, 0, 0, ptr, 0, NULL, &readEvent);
	ret |= (*clWaitForEvents) (1, &readEvent);
	ret |= (*clReleaseEvent) (readEvent);
	return ret;
}

int getBufferData(cl_mem buf, size_t size, void *ptr)
{
	void *event;
	cl_int ret = (*clEnqueueReadBuffer) (queue, buf, 1, 0, size, ptr, 0, NULL, &event);
	ret |= (*clWaitForEvents) (1, &event);
	ret |= (*clReleaseEvent) (event);
	return ret;
}

int setBufferData(cl_mem buf, size_t size, void *data)
{
	void *event;
	cl_int ret = (*clEnqueueWriteBuffer) (queue, buf, 1, 0, size, data, 0, NULL, &event);
	ret |= (*clWaitForEvents) (1, &event);
	ret |= (*clReleaseEvent) (event);
	return ret;
}

int assign_to_centers_kernel(imageResizer * ir, cl_mem out, unsigned int outWidth, unsigned int outHeight,
			     cl_mem centers, unsigned int k, cl_mem assignments, cl_mem changed)
{
	cl_int ret;
	ret = (*clSetKernelArg) (ir->assignToCentersKernel, 0, sizeof(cl_mem), &out);
	ret |= (*clSetKernelArg) (ir->assignToCentersKernel, 1, sizeof(unsigned int), &outWidth);
	ret |= (*clSetKernelArg) (ir->assignToCentersKernel, 2, sizeof(cl_mem), &centers);
	ret |= (*clSetKernelArg) (ir->assignToCentersKernel, 3, sizeof(unsigned int), &k);
	ret |= (*clSetKernelArg) (ir->assignToCentersKernel, 4, sizeof(cl_mem), &assignments);
	ret |= (*clSetKernelArg) (ir->assignToCentersKernel, 5, sizeof(cl_mem), &changed);
	if (ret != 0)
		return ret;

	void *event;
	size_t globalSize[] = { outWidth, outHeight };
	ret |= (*clEnqueueNDRangeKernel) (queue, ir->assignToCentersKernel, 2, 0, globalSize, 0, 0, NULL, &event);
	ret |= (*clWaitForEvents) (1, &event);
	ret |= (*clReleaseEvent) (event);
	return ret;
}

int recompute_centers(imageResizer * ir, cl_mem out, unsigned int outWidth, unsigned int outHeight, cl_mem centers,
		      unsigned int k, cl_mem assignments, cl_mem global_sum, cl_mem global_count)
{
	cl_int ret;
	ret = (*clSetKernelArg) (ir->recomputeCentersKernel, 0, sizeof(cl_mem), &out);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 1, sizeof(cl_uint), &outWidth);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 2, sizeof(cl_uint), &outHeight);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 3, sizeof(cl_mem), &centers);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 4, sizeof(cl_mem), &global_sum);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 5, sizeof(cl_mem), &global_count);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 6, sizeof(cl_mem), &assignments);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 7, k * sizeof(float) * 4, NULL);
	ret |= (*clSetKernelArg) (ir->recomputeCentersKernel, 8, k * sizeof(cl_uint), NULL);
	if (ret != 0)
		return ret;
	int window = 8;
	size_t globalSize = (outWidth * outHeight + 1) / window;
	size_t localSize = k;
	void *event;
	ret = (*clEnqueueNDRangeKernel) (queue, ir->recomputeCentersKernel, 1,	// work_dim
					 0,	// global_work_offset
					 &globalSize,	// global_work_size
					 &localSize,	// local_work_size
					 0,	// num_events_in_wait_list
					 NULL,	// event_wait_list
					 &event	//event
	    );
	ret |= (*clWaitForEvents) (1, &event);
	ret |= (*clReleaseEvent) (event);
	if (ret != 0)
		return ret;
}

int random_centers_from_image(imageResizer * ir, cl_mem out, unsigned int outWidth, unsigned int outHeight,
			      cl_mem centers, unsigned int k)
{
	cl_int ret;
	ret |= (*clSetKernelArg) (ir->randomCentersFromImageKernel, 0, sizeof(cl_mem), &out);
	ret |= (*clSetKernelArg) (ir->randomCentersFromImageKernel, 1, sizeof(cl_uint), &outWidth);
	ret |= (*clSetKernelArg) (ir->randomCentersFromImageKernel, 2, sizeof(cl_uint), &outHeight);
	ret |= (*clSetKernelArg) (ir->randomCentersFromImageKernel, 3, sizeof(cl_mem), &centers);
	unsigned int seed = rand();
	ret |= (*clSetKernelArg) (ir->randomCentersFromImageKernel, 4, sizeof(cl_uint), &seed);
	if (ret != 0)
		return ret;

	size_t globalSize = k;
	void *event;
	ret = (*clEnqueueNDRangeKernel) (queue, ir->randomCentersFromImageKernel, 1,	// work_dim
					 0,	// global_work_offset
					 &globalSize,	// global_work_size
					 0,	// local_work_size
					 0,	// num_events_in_wait_list
					 NULL,	// event_wait_list
					 &event	//event
	    );
	ret |= (*clWaitForEvents) (1, &event);
	ret |= (*clReleaseEvent) (event);
	return ret;
}

int convert_to_uint8(imageResizer * ir, cl_mem centers, unsigned int k, void *palette)
{
	cl_int ret;
	cl_mem paletteBuf = (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, k * 4 * sizeof(cl_uchar), NULL, &ret);
	if (ret != 0)
		return ret;
	ret = (*clSetKernelArg) (ir->convertToUint8Kernel, 0, sizeof(cl_mem), &centers);
	ret |= (*clSetKernelArg) (ir->convertToUint8Kernel, 1, sizeof(cl_mem), &paletteBuf);
	if (ret != 0)
		return ret;
	size_t globalSize = k;
	void *convertEvent;
	ret = (*clEnqueueNDRangeKernel) (queue, ir->convertToUint8Kernel, 1, 0, &globalSize, 0, 0, NULL, &convertEvent);
	if (ret != 0)
		return ret;
	void *readEvent;
	ret =
	    (*clEnqueueReadBuffer) (queue, paletteBuf, 1, 0, k * 4 * sizeof(cl_uchar), palette, 1, &convertEvent,
				    &readEvent);
	ret |= (*clWaitForEvents) (1, &readEvent);
	ret |= (*clReleaseEvent) (convertEvent);
	ret |= (*clReleaseEvent) (readEvent);
	ret |= (*clReleaseMemObject) (paletteBuf);
	return ret;
}

int add_one(imageResizer * ir, cl_mem assignments, size_t size, void *outData)
{
	cl_int ret;
	ret = (*clSetKernelArg) (ir->addOneKernel, 0, sizeof(cl_mem), &assignments);
	if (ret != 0)
		return ret;
	void *addOneEvent;
	ret = (*clEnqueueNDRangeKernel) (queue, ir->addOneKernel, 1, 0, &size, 0, 0, NULL, &addOneEvent);
	if (ret != 0)
		return ret;
	void *readEvent;
	ret = (*clEnqueueReadBuffer) (queue, assignments, 1, 0, size, outData, 1, &addOneEvent, &readEvent);
	ret |= (*clWaitForEvents) (1, &readEvent);
	ret |= (*clReleaseEvent) (addOneEvent);
	ret |= (*clReleaseEvent) (readEvent);
	return ret;
}

int resize_paletted(void *irp, unsigned int outWidth, unsigned int outHeight, void *outData, void *palette,
		    unsigned int k)
{
	if (k == 0 || k > 255) {
		return -100;
	}
	//make k-1 clusters to that the 0th cluster will be transparent
	k -= 1;
	imageResizer *ir = (imageResizer *) irp;
	cl_mem out;
	void *resizeKernel;
	cl_int ret = _resize(ir, outWidth, outHeight, &out, &resizeKernel);
	if (ret != 0)
		return ret;
	//k-means quantization (create paletted image)

	//allocate buffers
	cl_mem centers = (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, k * 4 * sizeof(float), NULL, &ret);
	if (ret != 0)
		return ret;
	cl_mem assignments =
	    (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, outWidth * outHeight * sizeof(cl_uchar), NULL, &ret);
	if (ret != 0)
		return ret;
	cl_mem global_sum = (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, k * 4 * sizeof(cl_ulong), NULL, &ret);
	if (ret != 0)
		return ret;
	cl_mem global_count = (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, k * sizeof(cl_uint), NULL, &ret);
	if (ret != 0)
		return ret;
	cl_int changed = 0;
	cl_mem changedBuf = (*clCreateBuffer) (deviceCtx, CL_MEM_READ_WRITE, sizeof(cl_int), NULL, &ret);
	if (ret != 0)
		return ret;
	(*clWaitForEvents) (1, &resizeKernel);
	(*clReleaseEvent) (resizeKernel);
	//init with random centers from image
	ret = random_centers_from_image(ir, out, outWidth, outHeight, centers, k);
	if (ret != 0)
		return ret;
	//find clusters
	while (1) {
		changed = 0;
		ret = setBufferData(changedBuf, sizeof(cl_int), (void *)(&changed));
		if (ret != 0)
			return ret;
		ret = assign_to_centers_kernel(ir, out, outWidth, outHeight, centers, k, assignments, changedBuf);
		if (ret != 0)
			return ret;
		ret = getBufferData(changedBuf, sizeof(cl_int), (void *)(&changed));
		if (ret != 0)
			return ret;
		if (changed == 0)
			break;
		ret =
		    recompute_centers(ir, out, outWidth, outHeight, centers, k, assignments, global_sum, global_count);
		if (ret != 0)
			return ret;
	}
	//get data
	((cl_uchar *) palette)[0] = 0;
	((cl_uchar *) palette)[1] = 0;
	((cl_uchar *) palette)[2] = 0;
	((cl_uchar *) palette)[3] = 0;
	ret = convert_to_uint8(ir, centers, k, palette + (4 * sizeof(cl_uchar)));
	ret = add_one(ir, assignments, outWidth * outHeight, outData);
	//release buffers
	ret |= (*clReleaseMemObject) (changedBuf);
	ret |= (*clReleaseMemObject) (centers);
	ret |= (*clReleaseMemObject) (assignments);
	ret |= (*clReleaseMemObject) (global_sum);
	ret |= (*clReleaseMemObject) (global_count);
	ret |= (*clReleaseMemObject) (out);

	return ret;
}
