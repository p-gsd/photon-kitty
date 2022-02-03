--adds arrow keys to move the selectedCard
photon = require("photon")

photon.keybindings.add(photon.NormalState, "<right>", photon.selectedCard.moveRight)
photon.keybindings.add(photon.NormalState, "<left>", photon.selectedCard.moveLeft)
photon.keybindings.add(photon.NormalState, "<up>", photon.selectedCard.moveUp)
photon.keybindings.add(photon.NormalState, "<down>", photon.selectedCard.moveDown)
