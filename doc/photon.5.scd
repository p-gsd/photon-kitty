photon(5)

# NAME

photon - config file syntax

# SYNOPSIS

photon reads the config file and interprets it's content as url for rss/atom
feeds. By default running photon with no arguments will try to read
*~/.config/photon/config*

# SYNTAX

photon parses the config file by lines. Every line can be either a:

- URL: direct url to a rss/atom feed
- Comment: start with *#* and end with the line break
- Command: is a line with the prefix: *cmd://* and then the command which will be called by the shell

## EXAMPLE

```
#fsf news feed \
https://static.fsf.org/fsforg/rss/news.xml \
#itsfoss feed \
https://itsfoss.com/feed/ \
#golang blog feed \
https://blog.golang.org/feed.atom\?format\=xml \
#some command (ratt searching youtube for cat videos) \
cmd://ratt auto https://https://www.youtube.com/results --data-urlencode="search_query=cat" 
```

# SEE ALSO

_photon_(5)

# AUTHORS

Maintained by Vladimír Magyar <vladimir@mgyar.me>. Up-to-date sources can be
found at https://sr.ht/~ghost08/photon and bugs/patches can be submitted by
email to ~ghost08/photon@todo.sr.ht
