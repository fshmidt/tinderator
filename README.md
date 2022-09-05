# tinderator
A small pet project that increases the conversion of wasting time on Tinder.
I noticed a long time ago that very rarely the most beautiful and liked girls swipe back, and it’s not surprising, because on tinder they have an oversupply of thousands of guys, 100% of whom swiped them before you. They definitely have better photos, more money etc so it's really tough red ocean. What I also noticed is that people often write their Instagram in bio in various forms, and on Instagram they answer an order of magnitude more often.
This is how Tinderator was born. It consists of two parts:
- The go-part gets a bunch of screenshots from tinder/another dating app and parses nicknames on instagram from there. It writes them to a text file and a Postgres database. All this is packaged in a container and can be run on a remote server.
- The python part receives a list of nicknames from the firest stage and, imitating human behavior, likes random photos by these nicknames and sends them unusual emoticons as a dialogue initiating message.*
Hundreds of conversations per day can be initiated this way. Communication opening conversion (response message) is about 30%.
And then everything is in your hands). It's just for your connection in the very beggining.

*Python-part is now has working form on my pc and it perfectly works with .txt files from go-parcer, but I haven't finished docker compose for this part. I'll commit and push it soon add well.
