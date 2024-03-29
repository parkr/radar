package radar

type testDataContainer struct {
	oldStyleBody string
	newStyleBody string

	comment1Body string
	comment2Body string

	expectedOldItems []RadarItem
	expectedNewItems []RadarItem

	newIssueBody string
}

var testData = testDataContainer{
	oldStyleBody: `
	A new day! Here's what you have saved:

	## [*Previously:*](https://github.com/parkr/daily/issues/1886)

	- [ ] [Ankify: convert notes to Anki cards](https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn)
	- [ ] [WHERE TO WATCH BROADWAY ONLINE: THE THEATER LOVER’S GUIDE TO STREAMING](https://broadwaydirect.com/where-to-watch-musicals-online-the-musical-lovers-guide-to-streaming/)
	- [ ] [Coding Relic](https://codingrelic.geekhold.com/?m=1)
	- [ ] [“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau](https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s)
	- [ ] [“Parrot Style” graphic tee, pullover hoodie, onesie, tank, and pullover crewneck by deadpine. | Cotton Bureau](https://cottonbureau.com/products/parrot-style#/1941684/tee-men-standard-tee-white-100percent-cotton-s)
	- [ ] [The West Wing Weekly | Cotton Bureau](https://cottonbureau.com/stores/the-west-wing-weekly#/shop)
	- [ ] [Fools Crow - Wikipedia](https://en.m.wikipedia.org/wiki/Fools_Crow)
	- [ ] [A Little Life - Wikipedia](https://en.m.wikipedia.org/wiki/A_Little_Life)
	- [ ] [Archiving a website with wget](https://gist.github.com/mullnerz/9fff80593d6b442d5c1b)
	- [ ] [gildas-lormeau/SingleFile: Web Extension for Firefox/Chrome/MS Edge and CLI tool to save a faithful copy of an entire web page in a single HTML file](https://github.com/gildas-lormeau/SingleFile?utm_source=hackernewsletter&utm_medium=email&utm_term=fav#projects-using-singlefile)
	- [ ] [direnv/direnv: unclutter your .profile](https://github.com/direnv/direnv?utm_source=tldrnewsletter)
	- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
	- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
	- [ ] [dns.toys/main.go at master · knadh/dns.toys · GitHub](https://github.com/knadh/dns.toys/blob/master/cmd/dnstoys/main.go)
	- [ ] [graham-essays: 📚 Download the full collection of Paul Graham essays in EPUB, Kindle & Markdown for easy reading.](https://github.com/ofou/graham-essays)
	- [ ] [Humism Watches](https://humism.com/)
	- [ ] [Song Notes](https://jamesfunk.net/songs-notes)
	- [ ] [Read Hackernews on Kindle](https://ktool.io/hacker-news-to-kindle)
	- [ ] [Email | Learn Netdata](https://learn.netdata.cloud/docs/agent/health/notifications/email)
	- [ ] [Keynote: The Potential of Machine Learning for Hardware Design - Jeff Dean](https://m.youtube.com/watch?v=FraDFZ2t__A&t=269s&pp=2AGNApACAQ%3D%3D)
	- [ ] [My simple note-taking setup | Zettelkasten in Obsidian | Step-by-step guide - YouTube](https://m.youtube.com/watch?v=E6ySG7xYgjY)
	- [ ] [Why Retaining Walls Collapse - YouTube](https://m.youtube.com/watch?v=--DKkzWVh-E)
	- [ ] [Heat Pumps are Not Hard: Here's what it will take to start pumping - YouTube](https://m.youtube.com/watch?v=43XKfuptnik)
	- [ ] [Day-Date Day Wheels, Ethical Quandaries, Counterweights, And A Question You Shouldn't Ask - YouTube](https://m.youtube.com/watch?v=VAt8_ow91yI)
	- [x] [The Mechanical Apple Watch | Watchfinder & Co. - YouTube](https://m.youtube.com/watch?v=BiPYOZnLJYo)
	- [ ] [Sundar Pichai, CEO of Google and Alphabet - YouTube](https://m.youtube.com/watch?v=j9qGmO8Yy-Y)
	- [ ] [Adam Falkner reads "Kissing Your Shoulder Blade Is the Most Honest Thing I've Done This Week" - YouTube](https://m.youtube.com/watch?v=QjkHEWFoEkY)
	- [ ] [Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube](https://m.youtube.com/watch?v=ymqKPz5kRXE)
	- [ ] [Life is too short for dated CLI tools (Twitter thread)](https://mobile.twitter.com/amilajack/status/1479328649820000256)
	- [ ] [Overcast auf Twitter: „Want to join the Overcast beta group? TestFlight: https://t.co/SQ97C8KmA0 Slack group for feedback, bug reports, and feature discussion: https://t.co/mC7rGQ43f1 Scammers sometimes charge for these links. Please don’t fall for it! Overcast’s beta is always free.“ / Twitter](https://mobile.twitter.com/OvercastFM/status/1514597131587313664)
	- [ ] ["Shelter In Place" 5lb Bag : Ritual Coffee Roasters](https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/)

	/cc @parkr
	`,

	newStyleBody: `
	A new day! Here's what you have saved:

	## *Previously:*

	- [ ] [Ankify: convert notes to Anki cards](https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn)
	- [ ] [WHERE TO WATCH BROADWAY ONLINE: THE THEATER LOVER’S GUIDE TO STREAMING](https://broadwaydirect.com/where-to-watch-musicals-online-the-musical-lovers-guide-to-streaming/)
	- [ ] [Coding Relic](https://codingrelic.geekhold.com/?m=1)
	- [ ] [“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau](https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s)
	- [ ] [“Parrot Style” graphic tee, pullover hoodie, onesie, tank, and pullover crewneck by deadpine. | Cotton Bureau](https://cottonbureau.com/products/parrot-style#/1941684/tee-men-standard-tee-white-100percent-cotton-s)
	- [ ] [The West Wing Weekly | Cotton Bureau](https://cottonbureau.com/stores/the-west-wing-weekly#/shop)
	- [ ] [Fools Crow - Wikipedia](https://en.m.wikipedia.org/wiki/Fools_Crow)
	- [ ] [A Little Life - Wikipedia](https://en.m.wikipedia.org/wiki/A_Little_Life)
	- [ ] [Archiving a website with wget](https://gist.github.com/mullnerz/9fff80593d6b442d5c1b)
	- [ ] [gildas-lormeau/SingleFile: Web Extension for Firefox/Chrome/MS Edge and CLI tool to save a faithful copy of an entire web page in a single HTML file](https://github.com/gildas-lormeau/SingleFile?utm_source=hackernewsletter&utm_medium=email&utm_term=fav#projects-using-singlefile)
	- [ ] [direnv/direnv: unclutter your .profile](https://github.com/direnv/direnv?utm_source=tldrnewsletter)
	- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
	- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
	- [ ] [dns.toys/main.go at master · knadh/dns.toys · GitHub](https://github.com/knadh/dns.toys/blob/master/cmd/dnstoys/main.go)
	- [ ] [graham-essays: 📚 Download the full collection of Paul Graham essays in EPUB, Kindle & Markdown for easy reading.](https://github.com/ofou/graham-essays)
	- [ ] [Humism Watches](https://humism.com/)
	- [ ] [Song Notes](https://jamesfunk.net/songs-notes)
	- [ ] [Read Hackernews on Kindle](https://ktool.io/hacker-news-to-kindle)
	- [ ] [Email | Learn Netdata](https://learn.netdata.cloud/docs/agent/health/notifications/email)
	- [ ] [Keynote: The Potential of Machine Learning for Hardware Design - Jeff Dean](https://m.youtube.com/watch?v=FraDFZ2t__A&t=269s&pp=2AGNApACAQ%3D%3D)
	- [ ] [My simple note-taking setup | Zettelkasten in Obsidian | Step-by-step guide - YouTube](https://m.youtube.com/watch?v=E6ySG7xYgjY)
	- [ ] [Why Retaining Walls Collapse - YouTube](https://m.youtube.com/watch?v=--DKkzWVh-E)
	- [ ] [Heat Pumps are Not Hard: Here's what it will take to start pumping - YouTube](https://m.youtube.com/watch?v=43XKfuptnik)
	- [ ] [Day-Date Day Wheels, Ethical Quandaries, Counterweights, And A Question You Shouldn't Ask - YouTube](https://m.youtube.com/watch?v=VAt8_ow91yI)
	- [x] [The Mechanical Apple Watch | Watchfinder & Co. - YouTube](https://m.youtube.com/watch?v=BiPYOZnLJYo)
	- [ ] [Sundar Pichai, CEO of Google and Alphabet - YouTube](https://m.youtube.com/watch?v=j9qGmO8Yy-Y)
	- [ ] [Adam Falkner reads "Kissing Your Shoulder Blade Is the Most Honest Thing I've Done This Week" - YouTube](https://m.youtube.com/watch?v=QjkHEWFoEkY)
	- [ ] [Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube](https://m.youtube.com/watch?v=ymqKPz5kRXE)
	- [ ] [Life is too short for dated CLI tools (Twitter thread)](https://mobile.twitter.com/amilajack/status/1479328649820000256)
	- [ ] [Overcast auf Twitter: „Want to join the Overcast beta group? TestFlight: https://t.co/SQ97C8KmA0 Slack group for feedback, bug reports, and feature discussion: https://t.co/mC7rGQ43f1 Scammers sometimes charge for these links. Please don’t fall for it! Overcast’s beta is always free.“ / Twitter](https://mobile.twitter.com/OvercastFM/status/1514597131587313664)
	- [ ] ["Shelter In Place" 5lb Bag : Ritual Coffee Roasters](https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/)

	*Previously:* https://github.com/parkr/daily/issues/1886

	/cc @parkr
	`,

	comment1Body: `- [ ] [Belkin Magnetic Fitness Mount - Apple](https://www.apple.com/shop/product/HPT82ZM/A/belkin-magnetic-fitness-mount)`,

	comment2Body: `- [ ] [HYROX](https://hyrox.com/the-fitness-race/)
- [X] [Example 1](https://example.com/1?)
- [ ] [Example 2](https://example.com/2?)`,

	expectedOldItems: []RadarItem{
		{Title: `Ankify: convert notes to Anki cards`, URL: `https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn`},
		{Title: `WHERE TO WATCH BROADWAY ONLINE: THE THEATER LOVER’S GUIDE TO STREAMING`, URL: `https://broadwaydirect.com/where-to-watch-musicals-online-the-musical-lovers-guide-to-streaming/`},
		{Title: `Coding Relic`, URL: `https://codingrelic.geekhold.com/?m=1`},
		{Title: `“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau`, URL: `https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s`},
		{Title: `“Parrot Style” graphic tee, pullover hoodie, onesie, tank, and pullover crewneck by deadpine. | Cotton Bureau`, URL: `https://cottonbureau.com/products/parrot-style#/1941684/tee-men-standard-tee-white-100percent-cotton-s`},
		{Title: `The West Wing Weekly | Cotton Bureau`, URL: `https://cottonbureau.com/stores/the-west-wing-weekly#/shop`},
		{Title: `Fools Crow - Wikipedia`, URL: `https://en.m.wikipedia.org/wiki/Fools_Crow`},
		{Title: `A Little Life - Wikipedia`, URL: `https://en.m.wikipedia.org/wiki/A_Little_Life`},
		{Title: `Archiving a website with wget`, URL: `https://gist.github.com/mullnerz/9fff80593d6b442d5c1b`},
		{Title: `gildas-lormeau/SingleFile: Web Extension for Firefox/Chrome/MS Edge and CLI tool to save a faithful copy of an entire web page in a single HTML file`, URL: `https://github.com/gildas-lormeau/SingleFile?utm_source=hackernewsletter&utm_medium=email&utm_term=fav#projects-using-singlefile`},
		{Title: `direnv/direnv: unclutter your .profile`, URL: `https://github.com/direnv/direnv?utm_source=tldrnewsletter`},
		{Title: `onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.`, URL: `https://github.com/onceupon/Bash-Oneliner`},
		{Title: `onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.`, URL: `https://github.com/onceupon/Bash-Oneliner`},
		{Title: `dns.toys/main.go at master · knadh/dns.toys · GitHub`, URL: `https://github.com/knadh/dns.toys/blob/master/cmd/dnstoys/main.go`},
		{Title: `graham-essays: 📚 Download the full collection of Paul Graham essays in EPUB, Kindle & Markdown for easy reading.`, URL: `https://github.com/ofou/graham-essays`},
		{Title: `Humism Watches`, URL: `https://humism.com/`},
		{Title: `Song Notes`, URL: `https://jamesfunk.net/songs-notes`},
		{Title: `Read Hackernews on Kindle`, URL: `https://ktool.io/hacker-news-to-kindle`},
		{Title: `Email | Learn Netdata`, URL: `https://learn.netdata.cloud/docs/agent/health/notifications/email`},
		{Title: `Keynote: The Potential of Machine Learning for Hardware Design - Jeff Dean`, URL: `https://m.youtube.com/watch?v=FraDFZ2t__A&t=269s&pp=2AGNApACAQ%3D%3D`},
		{Title: `My simple note-taking setup | Zettelkasten in Obsidian | Step-by-step guide - YouTube`, URL: `https://m.youtube.com/watch?v=E6ySG7xYgjY`},
		{Title: `Why Retaining Walls Collapse - YouTube`, URL: `https://m.youtube.com/watch?v=--DKkzWVh-E`},
		{Title: `Heat Pumps are Not Hard: Here's what it will take to start pumping - YouTube`, URL: `https://m.youtube.com/watch?v=43XKfuptnik`},
		{Title: `Day-Date Day Wheels, Ethical Quandaries, Counterweights, And A Question You Shouldn't Ask - YouTube`, URL: `https://m.youtube.com/watch?v=VAt8_ow91yI`},
		{Title: `Sundar Pichai, CEO of Google and Alphabet - YouTube`, URL: `https://m.youtube.com/watch?v=j9qGmO8Yy-Y`},
		{Title: `Adam Falkner reads "Kissing Your Shoulder Blade Is the Most Honest Thing I've Done This Week" - YouTube`, URL: `https://m.youtube.com/watch?v=QjkHEWFoEkY`},
		{Title: `Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube`, URL: `https://m.youtube.com/watch?v=ymqKPz5kRXE`},
		{Title: `Life is too short for dated CLI tools (Twitter thread)`, URL: `https://mobile.twitter.com/amilajack/status/1479328649820000256`},
		{Title: `Overcast auf Twitter: „Want to join the Overcast beta group? TestFlight: https://t.co/SQ97C8KmA0 Slack group for feedback, bug reports, and feature discussion: https://t.co/mC7rGQ43f1 Scammers sometimes charge for these links. Please don’t fall for it! Overcast’s beta is always free.“ / Twitter`, URL: `https://mobile.twitter.com/OvercastFM/status/1514597131587313664`},
		{Title: `"Shelter In Place" 5lb Bag : Ritual Coffee Roasters`, URL: `https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/`},
	},

	expectedNewItems: []RadarItem{
		{Title: `Belkin Magnetic Fitness Mount - Apple`, URL: `https://www.apple.com/shop/product/HPT82ZM/A/belkin-magnetic-fitness-mount`},
		{Title: `HYROX`, URL: `https://hyrox.com/the-fitness-race/`},
		{Title: `Example 2`, URL: `https://example.com/2?`},
	},

	newIssueBody: `A new day, @monalisa! Here's what you have saved:

## New:

  * [ ] [Example 2](https://example.com/2?)
  * [ ] [HYROX](https://hyrox.com/the-fitness-race/)
  * [ ] [Belkin Magnetic Fitness Mount - Apple](https://www.apple.com/shop/product/HPT82ZM/A/belkin-magnetic-fitness-mount)

## *Previously:*

  * [ ] [Ankify: convert notes to Anki cards](https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn)
  * [ ] [WHERE TO WATCH BROADWAY ONLINE: THE THEATER LOVER’S GUIDE TO STREAMING](https://broadwaydirect.com/where-to-watch-musicals-online-the-musical-lovers-guide-to-streaming/)
  * [ ] [Coding Relic](https://codingrelic.geekhold.com/?m=1)
  * [ ] [“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau](https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s)
  * [ ] [“Parrot Style” graphic tee, pullover hoodie, onesie, tank, and pullover crewneck by deadpine. | Cotton Bureau](https://cottonbureau.com/products/parrot-style#/1941684/tee-men-standard-tee-white-100percent-cotton-s)
  * [ ] [The West Wing Weekly | Cotton Bureau](https://cottonbureau.com/stores/the-west-wing-weekly#/shop)
  * [ ] [Fools Crow - Wikipedia](https://en.m.wikipedia.org/wiki/Fools_Crow)
  * [ ] [A Little Life - Wikipedia](https://en.m.wikipedia.org/wiki/A_Little_Life)
  * [ ] [Archiving a website with wget](https://gist.github.com/mullnerz/9fff80593d6b442d5c1b)
  * [ ] [gildas-lormeau/SingleFile: Web Extension for Firefox/Chrome/MS Edge and CLI tool to save a faithful copy of an entire web page in a single HTML file](https://github.com/gildas-lormeau/SingleFile?utm_source=hackernewsletter&utm_medium=email&utm_term=fav#projects-using-singlefile)
  * [ ] [direnv/direnv: unclutter your .profile](https://github.com/direnv/direnv?utm_source=tldrnewsletter)
  * [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
  * [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
  * [ ] [dns.toys/main.go at master · knadh/dns.toys · GitHub](https://github.com/knadh/dns.toys/blob/master/cmd/dnstoys/main.go)
  * [ ] [graham-essays: 📚 Download the full collection of Paul Graham essays in EPUB, Kindle & Markdown for easy reading.](https://github.com/ofou/graham-essays)
  * [ ] [Humism Watches](https://humism.com/)
  * [ ] [Song Notes](https://jamesfunk.net/songs-notes)
  * [ ] [Read Hackernews on Kindle](https://ktool.io/hacker-news-to-kindle)
  * [ ] [Email | Learn Netdata](https://learn.netdata.cloud/docs/agent/health/notifications/email)
  * [ ] [Keynote: The Potential of Machine Learning for Hardware Design - Jeff Dean](https://m.youtube.com/watch?v=FraDFZ2t__A&t=269s&pp=2AGNApACAQ%3D%3D)
  * [ ] [My simple note-taking setup | Zettelkasten in Obsidian | Step-by-step guide - YouTube](https://m.youtube.com/watch?v=E6ySG7xYgjY)
  * [ ] [Why Retaining Walls Collapse - YouTube](https://m.youtube.com/watch?v=--DKkzWVh-E)
  * [ ] [Heat Pumps are Not Hard: Here's what it will take to start pumping - YouTube](https://m.youtube.com/watch?v=43XKfuptnik)
  * [ ] [Day-Date Day Wheels, Ethical Quandaries, Counterweights, And A Question You Shouldn't Ask - YouTube](https://m.youtube.com/watch?v=VAt8_ow91yI)
  * [ ] [Sundar Pichai, CEO of Google and Alphabet - YouTube](https://m.youtube.com/watch?v=j9qGmO8Yy-Y)
  * [ ] [Adam Falkner reads "Kissing Your Shoulder Blade Is the Most Honest Thing I've Done This Week" - YouTube](https://m.youtube.com/watch?v=QjkHEWFoEkY)
  * [ ] [Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube](https://m.youtube.com/watch?v=ymqKPz5kRXE)
  * [ ] [Life is too short for dated CLI tools (Twitter thread)](https://mobile.twitter.com/amilajack/status/1479328649820000256)
  * [ ] [Overcast auf Twitter: „Want to join the Overcast beta group? TestFlight: https://t.co/SQ97C8KmA0 Slack group for feedback, bug reports, and feature discussion: https://t.co/mC7rGQ43f1 Scammers sometimes charge for these links. Please don’t fall for it! Overcast’s beta is always free.“ / Twitter](https://mobile.twitter.com/OvercastFM/status/1514597131587313664)
  * [ ] ["Shelter In Place" 5lb Bag : Ritual Coffee Roasters](https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/)

*Previously:* https://github.com/parkr-test/radar-test/issues/1887
`,
}
