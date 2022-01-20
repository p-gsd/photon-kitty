--translate-url is a extension that translates known input urls to rss urls
--now it works on: youtube channels, subreddits, odysee channels
photon = require("photon")

photon.events.subscribe(photon.events.Init, function(e)
	for i = 1, photon.feedInputs.len(), 1 do
		feed = photon.feedInputs.get(i)
		translator = match(feed)
		if translator ~= nil then
			photon.feedInputs.set(i, translator(feed))
		end
	end
end)

function match(feed)
	if feed:match("https://www.youtube.com/channel/.*") ~= nil then
		return ytchannelTranslator
	end
	if feed:match("https://www.youtube.com/user/.*") ~= nil then
		return ytuserTranslator
	end
	if feed:match("https://www.reddit.com/r/.*") ~= nil then
		return redditTranslator
	end
	if feed:match("https://odysee.com/.*") ~= nil then
		return odyseeTranslator
	end
	return nil
end

function ytchannelTranslator(feed)
	channelID = feed:match("https://www.youtube.com/channel/(.*)")
	return "https://www.youtube.com/feeds/videos.xml?channel_id=" .. channelID
end

function ytuserTranslator(feed)
	username = feed:match("https://www.youtube.com/user/(.*)")
	return "https://www.youtube.com/feeds/videos.xml?user=" .. username
end

function redditTranslator(feed)
	if feed:sub(#feed-3) == ".rss" then
		return feed
	end
	return feed .. ".rss"
end

function odyseeTranslator(feed)
	channelID = feed:match("https://odysee.com/(.*)")
	return "https://lbryfeed.melroy.org/channel/odysee/" .. channelID
end
