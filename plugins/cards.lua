photon = require("photon")

photon.events.subscribe(
	photon.events.FeedsDownloaded,
	function(e)
		print(photon.selectedCard:moveRight())
		print(photon.selectedCard:card():link())
	end
)
