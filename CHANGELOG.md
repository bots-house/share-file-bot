#  (2020-12-10)


### Bug Fixes

* **bot:** dots in title ([7155ddb](https://github.com/bots-house/share-file-bot/commit/7155ddb7953c2c80c4ea77f363ea658ce4e67261))
* **bot:** file not found, not produce error ([#46](https://github.com/bots-house/share-file-bot/issues/46)) ([033ca4e](https://github.com/bots-house/share-file-bot/commit/033ca4eed997832cb026e65edd0a6159477a4375))
* **bot:** https://sentry.io/share/issue/5ac9224b4c6e4dd4a53832ff1da9f3cd/ ([6f94d79](https://github.com/bots-house/share-file-bot/commit/6f94d79d083b6641159ae34af0f6c09e9c642301))
* **bot:** is not member, no chat reply in chats ([8c2db5e](https://github.com/bots-house/share-file-bot/commit/8c2db5e2d207d7ab0872c5d1c9931827ebeeb777))
* **bot:** misspell in alert ([66901eb](https://github.com/bots-house/share-file-bot/commit/66901eb0548fb053e1d0849329fe71df9aac8d59))
* **bot:** owned file linked post text ([a0fc106](https://github.com/bots-house/share-file-bot/commit/a0fc106541ea6dc72d96774f1954cec493d89557))
* **bot:** remove typing... ([1380dfb](https://github.com/bots-house/share-file-bot/commit/1380dfbd690858d177dacda0c234c4220a799a6d)), closes [#15](https://github.com/bots-house/share-file-bot/issues/15)
* **bot:** reset state on start ([fb4b2bc](https://github.com/bots-house/share-file-bot/commit/fb4b2bc916666a7b771cce4a1c258c99cb521981))
* **bot:** texts ([862ff3c](https://github.com/bots-house/share-file-bot/commit/862ff3c123449f92039d4ce3e11adc47f18e6cbe))
* **chats:** handle chat already exists ([#49](https://github.com/bots-house/share-file-bot/issues/49)) ([f9a0ef9](https://github.com/bots-house/share-file-bot/commit/f9a0ef96bdac67a89a7210044a53357366e127b7))
* **ci:** missing release-it config ([4e47eeb](https://github.com/bots-house/share-file-bot/commit/4e47eeba6e9e246b539588d028e1e6c622504477))
* **service/file:** check invite restriction ([9be8eb2](https://github.com/bots-house/share-file-bot/commit/9be8eb297435d03bb6c9114ea0ba0591e813b2c5))
* **store/postgres:** errors.As bug ([bf44e5c](https://github.com/bots-house/share-file-bot/commit/bf44e5c6ca03100c6ed68298165a7c9304e01afe))
* small bugs ([#64](https://github.com/bots-house/share-file-bot/issues/64)) ([d78ccaa](https://github.com/bots-house/share-file-bot/commit/d78ccaa7a495005e423dcdb7d8e3b95d6f9109d6))
* **ci:** revision build arg ([03ffad1](https://github.com/bots-house/share-file-bot/commit/03ffad145c61048b653c5db6512a50492991bc19))
* **store:** problems with pgbouncer in tx mode ([#63](https://github.com/bots-house/share-file-bot/issues/63)) ([4eda5bd](https://github.com/bots-house/share-file-bot/commit/4eda5bd1612b7ae3a8013adebb3bcecca1130c13))
* don't work in group chats ([7e3a60f](https://github.com/bots-house/share-file-bot/commit/7e3a60f342f6c0aa4855aef2c58cdc7221da119b))
* linter ([8a846a2](https://github.com/bots-house/share-file-bot/commit/8a846a239e95a2e54413f7eb6930051c0c2be30a))
* **store/chat:** count method returns user count ([eb690b9](https://github.com/bots-house/share-file-bot/commit/eb690b9257084be39d993742d3ba3aa9dd06394d))
* **store/postgres:** strange null ([d4b8b4b](https://github.com/bots-house/share-file-bot/commit/d4b8b4b887ac6188fd726b6d33e0a3f5b6d94570))
* bot middleware ([8a82ca0](https://github.com/bots-house/share-file-bot/commit/8a82ca074954c8f3c91dad62304bb5652e3ddc92))
* entity parse and sentry context  ([#54](https://github.com/bots-house/share-file-bot/issues/54)) ([286b7d5](https://github.com/bots-house/share-file-bot/commit/286b7d5830322fb69ca11267981fc7a6e0bb3b98))


### Features

* **ci:** deploy via releases ([#73](https://github.com/bots-house/share-file-bot/issues/73)) ([f1e4305](https://github.com/bots-house/share-file-bot/commit/f1e4305be1bc36823c00c18c67bea8c7613b7eba))
* linked post ([#65](https://github.com/bots-house/share-file-bot/issues/65)) ([ebff55e](https://github.com/bots-house/share-file-bot/commit/ebff55e09c277a84838ba7c8bb9056fbdf052d01))
* **admin:** add /admin command ([b0e95ac](https://github.com/bots-house/share-file-bot/commit/b0e95ac62bcd200813ee0794883fa32085a88264))
* **bot:** add ref tracking ([#39](https://github.com/bots-house/share-file-bot/issues/39)) ([8ba9601](https://github.com/bots-house/share-file-bot/commit/8ba960165246a550a9a692c4cfd7e42c1bbde555))
* **bot:** add refersh button for owners ([aa3ca99](https://github.com/bots-house/share-file-bot/commit/aa3ca998312a49b76938e4304c1d1d024ba1682c))
* **bot:** change start text, add about button ([bdf4e6b](https://github.com/bots-house/share-file-bot/commit/bdf4e6bdcf312d6e55e30229e370698038a375f2))
* **bot:** delete document ([d0613e0](https://github.com/bots-house/share-file-bot/commit/d0613e08116bad666926b1e78eab9484ef201526))
* **bot:** delete user upload message ([711ff82](https://github.com/bots-house/share-file-bot/commit/711ff82076bf87304ce5c01b3f8373b4cd4274a8))
* **bot:** don't return error to tg ([095da40](https://github.com/bots-house/share-file-bot/commit/095da40c8cf54b6f60ae18368b61eab45d192043))
* **bot:** execute async not required calls ([187d783](https://github.com/bots-house/share-file-bot/commit/187d78344dd5b25e56708f5c0904e65482eee803))
* **bot:** fast fix for copyright ([41551e9](https://github.com/bots-house/share-file-bot/commit/41551e99506addaeefd30c0152f91f91a18d26f1))
* **bot:** markdown escape ([06312b4](https://github.com/bots-house/share-file-bot/commit/06312b45c758f584a1dd90144e5185deda27dee3)), closes [#40](https://github.com/bots-house/share-file-bot/issues/40)
* **bot:** parse user input ([#59](https://github.com/bots-house/share-file-bot/issues/59)) ([ae55ce5](https://github.com/bots-house/share-file-bot/commit/ae55ce5d43d5fee7d748e35d6fe5c8431e11dd7f))
* **bot:** refs with file ([#53](https://github.com/bots-house/share-file-bot/issues/53)) ([3b876f3](https://github.com/bots-house/share-file-bot/commit/3b876f397ed92cab7facba43a47ea4b61fd4a417))
* **bot:** temporary disable audio ([7df94e5](https://github.com/bots-house/share-file-bot/commit/7df94e57bbffe270445ff93578888ed71396a461))
* **main:** add revision to sentry ([eac83f6](https://github.com/bots-house/share-file-bot/commit/eac83f6cc1d547b0b1db340f3838cfc46912a7f0))
* **service/chat:** handle bot is not admin ([d27dae8](https://github.com/bots-house/share-file-bot/commit/d27dae8e21c6ec494a76889d5f8330db01634690))
* docker support  ([#31](https://github.com/bots-house/share-file-bot/issues/31)) ([aa71306](https://github.com/bots-house/share-file-bot/commit/aa71306f59252fad71c0cf216803dd8239b0a31c))
* request subscription to group/channel for download files ([#35](https://github.com/bots-house/share-file-bot/issues/35)) ([8353fb6](https://github.com/bots-house/share-file-bot/commit/8353fb6127d271466feae9189849ed85233dcb22))
* settings and long ids ([#25](https://github.com/bots-house/share-file-bot/issues/25)) ([4c0f05c](https://github.com/bots-house/share-file-bot/commit/4c0f05c458a850687df9a9348a02378bb7a3631b))
* support all content types  ([#33](https://github.com/bots-house/share-file-bot/issues/33)) ([8a0ab4a](https://github.com/bots-house/share-file-bot/commit/8a0ab4a566c9407b06b6265a759e4917fa4dd3c9))


### Reverts

* Revert "ci(deploy): remove old package versions" ([0b389a8](https://github.com/bots-house/share-file-bot/commit/0b389a8e27d9f33dc8150ed92983357168c6b374))

