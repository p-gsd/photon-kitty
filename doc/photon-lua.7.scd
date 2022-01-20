photon-lua(7)

# NAME

photon-lua - lua plugins documentation and tutorial for *photon*(1)

# SYNOPSIS

photon can be extended trough lua plugins which are automagically loaded from
*~/.config/photon/plugins/\*.lua*

# DESCRIPTION

Lua plugins can subscribe for events that will then call a lua callback
function. Generally they then change some state and exit.

Hello world example:

~/.config/photon/plugins/hello_world.lua

```
photon = require("photon")

photon.events.subscribe(
	photon.events.Init,
	function()
		print("Hello world!")
	end
)
```

# LIBRARY

The `photon` can be used to subscribe to events, get and set the state of
cards and so on.

## PHOTON

*state()*
	returns the actual state of the application
	it can be: *photon.Normal*, *photon.Article*, *photon.Search*

*cards*
	are all the loaded cards. see _CARDS_

*visibleCards*
	also a _CARDS_ object. Contains only visible cards, that are filtered.

*selectedCard*
	the selected card object.
	Methods:
		- *posX()* X position in the grid
		- *posY()* Y position in the grid
		- *card()* _CARD_ object

## CARDS

Array of _CARD_ object.

*len()*
	length of the cards array

*get(index)*
	get the _CARD_ object at *index*

*set(index, card)*
	set _CARD_ at *index*

*del(index)*
	delete _CARD_ at *index*

*add(index, card)*
	add a new _CARD_ at *index*

*append(card)*
	append a new _CARD_

*create(table)*
	returns a new _CARD_ object created from a lua table

```
newCard = {}
newCard["link"] = "https://example.com/article/2"
newCard["title"] = "article 2"
c = photon.cards.create(newCard)
```

## CARD

Object representing one card in the feed.

*link()*
	getter and setter for link

*image()*
	getter and setter for image

*title()*
	getter and setter for title

*content()*
	getter and setter for content

*description()*
	getter and setter for description

*published()*
	getter and setter for published

*feed()*
	gets the _FEED_ object of the _CARD_
	
*getMedia()*
	extracts the _MEDIA_ object

*runMedia()*
	runs the _MEDIA_. Opens a default application.

*openBrowser()*
	opens the _CARD_ link in the default browser/application.

*openArticle()*
	opens the _ARTICLE VIEW_ for the card.

All getters and setters can be used as so:

```
card = photon.cards.get(1)
--get the link
print("old link: " ... card:link())
--set the link to new value
card:link("https://newlink.com")
print("new link: " ... card:link())
```

## MEDIA

*originalLink*
	original link from the _CARD_, from which the media link is extracted.

*links*
	table of extracted media links. Usually with just one link.

*contentType*
	content type of the media.

*run()*
	runs the _MEDIA_. Opens a default application.

## EVENTS

*subscribe(eventType, function(event))*
	plugins can subscribe to a event type and a callback function which takes a
	event object.

Event types:

*Init*
	when photon has successfully started, and loaded the _INPUTS_ (either from
	the args or a config file)

*RunMediaStart*
	when a media link is about to play.

*RunMediaEnd*
	when a media link is done playing.

*FeedsDownloaded*
	when all the rss/atom feeds data is downloaded and parsed. _CARDS_ are from
	now accessible.

## INPUTS

*photon.inputs*
	an array of links/cmds that photon got from args, or parsed from a config
	file

It has the following functions:

*len()*
	returns the length of inputs

*get(index)*
	returns a link/cmd at *index*

*set(index, string)*
	sets a link at index

*add(index, string)*
	add a new link at index

*append(string)*
	the same as *inputs:add(inputs:len(), link)*

## KEYBINDINGS

TODO

# EXAMPLES

TODO

# SEE ALSO

_photon_(1)

# AUTHORS

Maintained by Vladimír Magyar <vladimir@mgyar.me>. Up-to-date sources can be
found at https://sr.ht/~ghost08/photon and bugs/patches can be submitted by
email to ~ghost08/photon@todo.sr.ht
