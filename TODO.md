## bugs

### card

-   add a tiny feed icon
-   check if there is place for the author name

### article

-   status bar - show percentage in article scroll, like in vim (Top-percent-Bot or All)
-   keybinding to toggle the card description or the article parsed content

### grid

-   ctrl + d/u/f/b
-   status bar - also show scroll percentage

```
if scrollHeight != viewportHeight
	pct = Math.round scrollTop / (scrollHeight - viewportHeight) * 100

	if isNaN(pct) then pct = 100
	if pct > 100  then pct = 100
	if pct < 0    then pct = 0
```

### status bar

-   show what feed is currently downloading
-   show media playing

### plugins

-   for reddit link type posts, opening the article with the link to the web page instead of the reddit post

### documentation

-   man pages (this needs then a makefile and scdoc)
-   better readme

### distribution

-   aur package

### ideas:

-   maybe to move libphoton internally
-   remove the movecard functions from callbacks and don't use them from lua
-   what if we don't use the build in article view, but just use a external tool, like less, bat, w3c and make the article view as a separate cli tool
