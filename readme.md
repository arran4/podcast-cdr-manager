# podcast-cdr-manager

CLI tool to help manage podcast subscriptions for burning to CDROMs / CDR / CDRW

The purpose this app is to manage CDs of Podcast in the form of MP3s. It will maintain a list of podcasts episodes which 
have  been played, create new ISOs with the next set of podcasts which haven't been listened to and update feeds you're
subscribed to.

# Usage

| Function                                                  | Command                                                                                                             |
|-----------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `podcastcdrmanager profile new default`                   | Creates a new profile (required)                                                                                    |
| `podcastcdrmanager subscribe rss http://`                 | Subscribes to a podcast RSS feed                                                                                    |
| `podcastcdrmanager disk next -create`                     | (Dry run) Creates a disk to start storing data on and fills it up with previously unused podcasts. (Or attempts to) |
| `podcastcdrmanager disk next -create -dry=false`          | Create a disk to store the data on and fills it up with previously unused podcasts. (Or attempts to)                |
| `podcastcdrmanager disk iso generate  -index 0`           | (Dry run) Generates an ISO from the contents                                                                        |
| `podcastcdrmanager disk iso generate -dry=false -index 0` | Generates an ISO from the contents                                                                                  |
| `podcastcdrmanager subscriptions list`                    | Lists podcasts                                                                                                      |
| `podcastcdrmanager subscriptions refresh`                 | Goes though podcast subscriptions and checks for updates adding the results to the unused podcast list              |
| `podcastcdrmanager disk list`                             | Lists disks                                                                                                         |
| `podcastcdrmanager cast list`                             | Lists podcast episodes.                                                                                             |

More to come on an as need basis.

# Example

```shell
% ./podcastcdrmanager profile new default-dev                                         
Profiled

% ./podcastcdrmanager subscribe rss http://localhost/\~arran/cache/podcast1/sounds.rss
Subscribed, added 17 new items

% ./podcastcdrmanager disk next -create
Put http://localhost/~arran/cache/podcast1/561424362-fallofcivilizations-roman-britain-the-work-of-giants-crumbled.mp3 on ability-faith-population.iso (1 mb + 59 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/562478415-fallofcivilizations-the-bronze-age-collapse-mediterranean-apocalypse.mp3 on ability-faith-population.iso (60 mb + 60 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/571604061-fallofcivilizations-3-the-mayan-collapse-ruins-among-the-trees.mp3 on ability-faith-population.iso (120 mb + 65 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/595152159-fallofcivilizations-4-the-greenland-vikings-land-of-the-midnight-sun.mp3 on ability-faith-population.iso (185 mb + 76 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/615124251-fallofcivilizations-5-the-khmer-empire-fall-of-the-god-kings.mp3 on ability-faith-population.iso (261 mb + 91 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/643783380-fallofcivilizations-6-easter-island-where-giants-walked.mp3 on ability-faith-population.iso (352 mb + 95 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/672619901-fallofcivilizations-7-the-songhai-empire-africas-age-of-gold.mp3 on ability-faith-population.iso (447 mb + 125 mb / 600 mb)
Ran out of space on ability-faith-population.iso (572 mb + 137 mb / 600 mb) (Meaning it's ready to generate an ISO from)
Dry not not saving to changes to profile

% ./podcastcdrmanager disk next -create -dry=false
Put http://localhost/~arran/cache/podcast1/561424362-fallofcivilizations-roman-britain-the-work-of-giants-crumbled.mp3 on ability-faith-population.iso (1 mb + 59 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/562478415-fallofcivilizations-the-bronze-age-collapse-mediterranean-apocalypse.mp3 on ability-faith-population.iso (60 mb + 60 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/571604061-fallofcivilizations-3-the-mayan-collapse-ruins-among-the-trees.mp3 on ability-faith-population.iso (120 mb + 65 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/595152159-fallofcivilizations-4-the-greenland-vikings-land-of-the-midnight-sun.mp3 on ability-faith-population.iso (185 mb + 76 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/615124251-fallofcivilizations-5-the-khmer-empire-fall-of-the-god-kings.mp3 on ability-faith-population.iso (261 mb + 91 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/643783380-fallofcivilizations-6-easter-island-where-giants-walked.mp3 on ability-faith-population.iso (352 mb + 95 mb / 600 mb)
Put http://localhost/~arran/cache/podcast1/672619901-fallofcivilizations-7-the-songhai-empire-africas-age-of-gold.mp3 on ability-faith-population.iso (447 mb + 125 mb / 600 mb)
Ran out of space on ability-faith-population.iso (572 mb + 137 mb / 600 mb) (Meaning it's ready to generate an ISO from)

% ./podcastcdrmanager disk iso generate -dry=false -index 0                           
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/672619901-fallofcivilizations-7-the-songhai-empire-africas-age-of-gold.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/561424362-fallofcivilizations-roman-britain-the-work-of-giants-crumbled.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/615124251-fallofcivilizations-5-the-khmer-empire-fall-of-the-god-kings.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/595152159-fallofcivilizations-4-the-greenland-vikings-land-of-the-midnight-sun.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/561424362-fallofcivilizations-roman-britain-the-work-of-giants-crumbled.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/562478415-fallofcivilizations-the-bronze-age-collapse-mediterranean-apocalypse.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/595152159-fallofcivilizations-4-the-greenland-vikings-land-of-the-midnight-sun.mp3
2023/10/01 18:19:14 Written: 561424362-fallofcivilizations-roman-britain-the-work-of-giants-crumbled.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/615124251-fallofcivilizations-5-the-khmer-empire-fall-of-the-god-kings.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/571604061-fallofcivilizations-3-the-mayan-collapse-ruins-among-the-trees.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/672619901-fallofcivilizations-7-the-songhai-empire-africas-age-of-gold.mp3
2023/10/01 18:19:14 Written: 595152159-fallofcivilizations-4-the-greenland-vikings-land-of-the-midnight-sun.mp3
2023/10/01 18:19:14 Downloading http://localhost/~arran/cache/podcast1/643783380-fallofcivilizations-6-easter-island-where-giants-walked.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/562478415-fallofcivilizations-the-bronze-age-collapse-mediterranean-apocalypse.mp3
2023/10/01 18:19:14 Written: 615124251-fallofcivilizations-5-the-khmer-empire-fall-of-the-god-kings.mp3
2023/10/01 18:19:14 Written: 672619901-fallofcivilizations-7-the-songhai-empire-africas-age-of-gold.mp3
2023/10/01 18:19:14 Downloaded http://localhost/~arran/cache/podcast1/571604061-fallofcivilizations-3-the-mayan-collapse-ruins-among-the-trees.mp3
2023/10/01 18:19:14 Written: 562478415-fallofcivilizations-the-bronze-age-collapse-mediterranean-apocalypse.mp3
2023/10/01 18:19:15 Downloaded http://localhost/~arran/cache/podcast1/643783380-fallofcivilizations-6-easter-island-where-giants-walked.mp3
2023/10/01 18:19:15 Written: 571604061-fallofcivilizations-3-the-mayan-collapse-ruins-among-the-trees.mp3
2023/10/01 18:19:15 Written: 643783380-fallofcivilizations-6-easter-island-where-giants-walked.mp3
ISO generated


% ls -lh *.iso
-rw-r--r-- 1 arran arran 568M Oct  1 18:19 ability-faith-population.iso

```

