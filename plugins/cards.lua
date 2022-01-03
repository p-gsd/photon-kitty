photon = require("photon")
events = require("photon.events")

events.subscribe(events.FeedsDownloaded, function(e)
	print(photon.selectedCard:moveRight())
	print(photon.selectedCard:card():link())
end)
