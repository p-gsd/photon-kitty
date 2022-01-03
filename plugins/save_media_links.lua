--this plugin gets the media of the selectedCard and saves the links to a text file
keybindings = require("photon.keybindings")

keybindings.add(photon.NormalState, "dl", function()
	local media, err = photon.selectedCard.getMedia()
	if err ~= nil then 
		error(err)
	end
	local f = io.open("/tmp/links.txt", "w")
	for _, link in ipairs(media.links) do
		f:write(link)
	end
	f:close()
end)
