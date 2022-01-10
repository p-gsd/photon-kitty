--reddit-link is a plugin, that checks if the reddit post links to a article, and if it does, replaces the item.link with it
photon = require("photon")
events = require("photon.events")

events.subscribe(events.FeedsDownloaded, function(e)
	for i = 1, photon.cards:len(), 1 do
		card = photon.cards:get(i)
		content = card:content()
		--find link in content
		s, e = content:find('<a href="[^"]+">%[link%]')
		if s ~= nil then
			--extract link from content
			link = content:sub(s+9, e-8)
			--if the link points somewhere else then reddit replace it
			if link:find("https://www.reddit.com/", 1, true) ~= 1 then
				card:link(link)
			end
		end
	end
end)
