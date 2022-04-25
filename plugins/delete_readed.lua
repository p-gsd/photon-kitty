--this plugin saves readed (opened/article opened/played) cards to localStorage and shows them in different color
photon = require("photon")
localStorage = require("localStorage")

photon.events.subscribe(
	photon.events.FeedsDownloaded,
	function(e)
		local delete = {}
		for i = 1, photon.cards:len(), 1 do
			local card = photon.cards:get(i)
			local item = localStorage.getItem(card:link()) 
			if item ~= nil then
				table.insert(delete, i)
			end
		end
		local deletedCount = 0
		for i = 1, #delete, 1 do
			photon.cards:del(delete[i] - deletedCount)
			deletedCount = deletedCount + 1 
		end
	end
)

function opened(e)
	if localStorage.getItem(e:link()) ~= nil then
		return
	end
	localStorage.setItem(e:link(), "")
	for i = 1, photon.cards:len(), 1 do
		local card = photon.cards:get(i)
		if card:link() == e:link() then
			photon.cards:del(i)
			return
		end
	end
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
	photon.events.RunMediaEnd,
	opened
)
