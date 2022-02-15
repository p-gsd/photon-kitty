### perf

-   sixelBuf -> SixelPrinter - an array of position + pointer to sixel data

```go
type SixelPrinter []struct{
	pos image.Point
	sixelData []byte
}
```

### documentation

-   lua man pages

### plugins

-   events.subscribe with priority

### status bar

-   prettier (maybe colors?)

### ideas:

-   extractor download cmd
-   add arg to use mercury-parser as article extractor?
