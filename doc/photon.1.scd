photon(1)

# NAME

photon - RSS reader as light as a photon with terminal + sixel

# SYNOPSIS

_photon_ [OPTION] [<paths> ...]

# DESCRIPTION

photon is a RSS/Atom reader with the focus on speed, usability and a bit of unix
philosophy. It shows the items in a card grid, with one selected card. User can
then scroll through it, read the article, play media or open in default
application.

# OPTIONS

*[<paths> ...]*
	RSS/Atom urls, config path, or - for _stdin_

*-h*, *--help*
	Show context-sensitive help.

*--extractor*
	command for media link extraction
	item link is substituted for %
	env: PHOTON_EXTRACTOR
	Default: *yt-dlp --get-url %*

*--video-cmd*
	set default command for opening the item media link in a video player
	media link is substituted for *%*
	direct item link is substituted for *$*
	if no *%* or *$* is provided, photon will download the data and pipe it to the stdin of the command 
	env: PHOTON_VIDEOCMD
	Default: *mpv $*

*--image-cmd*
	set default command for opening the item media link in a image viewer
	media link is substituted for *%*
	direct item link is substituted for *$*
	if no *%* or *$* is provided photon will download the data and pipe it to the stdin of the command
	env: PHOTON_IMAGECMD
	Default: *imv -*

*--torrent-cmd*
	set default command for opening the item media link in a torrent downloader
	media link is substituted for *%*
	if link is a torrent file, photon will download it, and substitute the
	torrent file path for *%*
	env: PHOTON_TORRENTCMD
	Default: *mpv %*

*--article-mode*
	the default article view mode
	env: PHOTON_ARTICLE_MODE
	Default: *ARTICLE*

*--article-renderer*
	command to render the item.Content/item.Description
	env: PHOTON_ARTICLE_RENDERER
	Default: *w3m -T text/html -dump -cols 72*

*-b*, *--cookie=KEY=VALUE;...*
	sets the cookie of all outgoing http requests

*-H*, *--header=KEY=VALUE;...*
	sets the headers of all outgoing http requests

*-A*, *--user-agent*
	sets the User-Agent header of all outgoing http request

*-k*, *--insecure*
	(TLS) By default, every SSL connection photon makes is verified to be
	secure. This option allows photon to proceed and operate even for server
	connections otherwise considered insecure

*--download-path*
	the default download path
	Default: *$HOME/Downloads*

*-t*, *--terminal-title*
	set the terminal title

# USAGE

## GENERAL

photon's philosophy is to be a some kind of rss viewer. So it doesn't have
features like adding or managing feeds. photon can be started with a list of
urls:

	$ photon https://static.fsf.org/fsforg/rss/news.xml https://itsfoss.com/feed/ https://blog.golang.org/feed.atom\?format\=xml

Or it can be directly fed with the feed data:

	$ curl -s https://www.reddit.com/r/pics.rss | photon -

Or using ratt and playing magnet links directly with mpv:

	$ ratt auto https://1337x.to/top-100 | photon -

Or searching youtube using _ratt_(1):

	$ ratt auto https://www.youtube.com/results --data-urlencode="search_query=MyQuery" | photon -

Or it can parse a config file (see _photon_(5)):

```
$ cat mychannels.txt
#youtube - LunusTechTips
https://www.youtube.com/feeds/videos.xml?user=LinusTechTips
#youtube - Rob Braxman Tech
https://www.youtube.com/feeds/videos.xml?channel_id=UCYVU6rModlGxvJbszCclGGw
#odysee - Lunduke on Tech" > mychannels.txt
https://lbryfeed.melroy.org/channel/odysee/@Lunduke:e
$ photon mychannels.txt
```

Or it can call a external command, must be specified with the *cmd://* prefix:

```
$ echo "cmd://ratt auto https://videoportal.joj.sk/slovania" >> mychannels.txt
$ photon mychannels.txt
```

By default running photon with no arguments will try to read
*~/.config/photon/config*

So the feed management is up to the user with arguments, text files and creating
scripts like:

```
$ cat photon-reddit.sh
photon mysubreddits.conf
$ cat mysubreddits.conf
https://www.reddit.com/r/popular.rss
https://www.reddit.com/r/news.rss
https://www.reddit.com/r/funny.rss
https://www.reddit.com/r/memes.rss
```

## ARTICLE VIEW

By pressing *ENTER*, photon will show the article view, where it scraps the
card's link and extracts the title, top image and main text content. The article
view also has three modes:

- *ARTICLE* - shows the scrapped article content

- *DESCRIPTION* - shows the item.Description

- *CONTENT* - shows the item.Content

Article view in *DESCRIPTION* or *CONTENT* mode, can use a external tool to
render the text. By default *w3m* is used, but can be changed with the
*--article-renderer* argument, or *PHOTON_ARTICLE_RENDERER* environment
variable.

## MEDIA EXTRACTION

photon can extract the direct media link of the rss item. Media extraction is by
default done by _yt-dlp_. By the content type of the media, photon will run it
in either a video player (default _mpv_) or a image viewer (default _imv_). If
the link is a magnet link, or a torrent file, photon will run it in a torrent
downloader/player (default _mpv_ with the _webtorrent-mpv-hook_ script).

# KEYBINDINGS

## CARD VIEW

*h* move selected card to the left

*j* move selected card down

*k* move selected card up

*l* move selected card to the right

*CTRL+-* will decrease the number of columns of cards.

*CTRL+=* will increase the number of columns of cards.

*ENTER*  will show the article view

*p* will play the media link

*o* will open the card's link in the default web browser (or default application).

*yy* - copy card link to clipboard

*yi* - copy the card image

*dm* - download media

*dl* - download link content

*di* - download image

*CTRL+d* - scroll half screen down

*CTRL+u* - scroll half screen up

*CTRL+f* - scroll full screen down

*CTRL+b* - scroll full screen up

*gg* - go to the first line

*G* - go to the last line

*q* - exit the application

## ARTICLE VIEW

*m* circle trough article view mode

*j* scroll the article down

k scroll the article up

*CTRL+d* scroll half screen down

*CTRL+u* scroll half screen up

*CTRL+f* scroll full screen down

*CTRL+b* scroll full screen up

*gg* scroll the article to the top

*G* scroll the article to the bottom.

*q* or *Esc* close the article.

## SEARCH

Searching is done with pressing */* and then typing the query. photon will
filter the visible cards by finding the query in either the card's title,
description, feed's title.

# LUA PLUGINS

photon will automatically load lua scripts from
*~/.config/photon/plugins/\*.lua*

These plugins help to extend the functionality and make photon do whatever the
user needs. Lua plugins can subscribe for events or register new keybindings and
modify the state. To see the entire photon lua plugins tutorial please see
_photon-lua_(7)

# SEE ALSO

_photon_(5), _photon-lua_(7)

# AUTHORS

Maintained by Vladimír Magyar <vladimir@mgyar.me>. Up-to-date sources can be
found at https://sr.ht/~ghost08/photon and bugs/patches can be submitted by
email to ~ghost08/photon@todo.sr.ht
