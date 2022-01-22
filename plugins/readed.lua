--this plugin saves readed (opened/article opened/played) cards to localStorage and shows them in different color
photon = require("photon")
localStorage = require("localStorage")

photon.events.subscribe(
	photon.events.FeedsDownloaded,
	function(e)
		for i = 1, photon.cards:len(), 1 do
			local card = photon.cards:get(i)
			local link = card:link()
			local item = localStorage.getItem(link) 
			if item ~= nil then
				card:foreground(photon.ColorPurple)
			end
		end
	end
)

function opened(e)
	if localStorage.getItem(e:link()) ~= nil then
		return
	end
	localStorage.setItem(e:link(), "")
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
