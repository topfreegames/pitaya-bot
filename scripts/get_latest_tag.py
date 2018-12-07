# -*- coding: utf-8 -*-
# Pitaya-Bot
# https://github.com/topfreegames/pitaya-bot
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

import urllib
import urllib2
import json

def main():
    url = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:tfgco/pitaya-bot:pull,push"
    response = urllib.urlopen(url)
    token = json.loads(response.read())['token']

    url = "https://registry-1.docker.io/v2/tfgco/pitaya-bot/tags/list"
    req = urllib2.Request(url, None, {
        "Authorization": "Bearer %s" % token,
    })
    response = urllib2.urlopen(req)
    tags = json.loads(response.read())
    last_tag = get_last_tag(tags['tags'])
    print last_tag


def get_last_tag(tags):
    valid_tags = filter(lambda t: t != 'latest', tags)
    return max([(t.split('-')[0], t) for t in valid_tags], key=lambda i: i[0])[1]


if __name__ == "__main__":
    main()
