--this plugin saves readed (opened/article opened/played) cards to a file and shows them in different color
photon = require("photon")
fs = require("fs")

home = fs.home()

photon.events.subscribe(
	photon.events.FeedsDownloaded,
	function(e)
		for i = 1, photon.cards:len(), 1 do
			local card = photon.cards:get(i)
			if readed(card:link()) then
				card:foreground(photon.ColorPurple)
			end
		end
	end
)

function opened(e)
	if readed(e:link()) then
		return
	end
	fs.mkdirAll(home .. "/.cache/photon")
	file = io.open(home .. "/.cache/photon/readed.data", "a+")
	file:write(e:link() .. "\n")
	file:close()
	e:card():foreground(photon.ColorPurple)
end

photon.events.subscribe(
	photon.events.ArticleOpened,
	opened
)

photon.events.subscribe(
	photon.events.LinkOpened,
	opened
)

photon.events.subscribe(
	photon.events.RunMediaStart,
	opened
)

function readed(link)
	fs.mkdirAll(home .. "/.cache/photon")
	file = io.open(home .. "/.cache/photon/readed.data", "r")
	if file == nil then
		return false
	end
	local line = file:read()
	while line ~= "" and line ~= nil do
		if link == line then
			return true
		end
		line = file:read()
	end
	file:close()
	return false
end
