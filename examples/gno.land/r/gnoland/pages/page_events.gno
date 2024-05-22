package gnopages

func init() {
	var (
		path  = "events"
		title = "Gno.land Core Team Attends Industry Events & Meetups"
		// XXX: description = "If you’re interested in learning more about Gno.land, you can join us at major blockchain industry events throughout the year either in person or virtually."
		body = `
If you’re interested in building web3 with us, catch up with Gno.land in person at one of our industry events.
We’re looking to connect with developers and like-minded thinkers who can contribute to the growth of our platform.

---

## Upcoming Events

<div class="columns-3">

<div class="column">

### Golang Serbia Meetup

- Join the meetup
- Belgrade, May 23, 2024

[Learn more](https://www.meetup.com/golang-serbia/events/300975006/)

</div><!-- end column-->

<div class="column">

### GopherCon EU
- Come Meet Us at our Booth
- Berlin, June 17 - 20, 2024

[Learn More](https://gophercon.eu/)
</div><!-- end column-->

<div class="column">

### GopherCon US
- Come Meet Us at our Booth
- Chicago, July 7 - 10, 2024

[Learn More](https://www.gophercon.com/)

</div><!-- end column-->

<div class="column">

### Nebular Summit
- Join our workshop
- Brussels, July 12 - 13, 2024

[Learn More](https://nebular.builders/)
</div><!-- end column-->

<div class="column">

</div><!-- end column-->
<div class="column">

</div><!-- end column-->
</div><!-- end columns-3-->

---

## Past Events

<div class="columns-3">
<div class="column">

### GopherCon US

- **Come Meet Us at our Booth**
- San Diego, September 26 - 29, 2023

[Learn more](https://www.gophercon.com/)

</div><!-- end column-->

<div class="column">

### GopherCon EU

- **Come Meet Us at our Booth**
- Berlin, July 26 - 29, 2023

[Learn more](https://gophercon.eu/)

</div><!-- end column-->

<div class="column">

### Nebular Summit Gno.land for Developers

- Paris, July 24 - 25, 2023
- Manfred Touron

[Learn more](https://www.nebular.builders/)

</div><!-- end column-->

<div class="column">

### EthCC

- **Come Meet Us at our Booth**
- Paris, July 17 - 20, 2023
- Manfred Touron

[Learn more](https://www.ethcc.io/)

</div><!-- end column-->

<div class="column">

### Eth Seoul

- **The Evolution of Smart Contracts: A Journey into Gno.land**
- Seoul, June 3, 2023
- Manfred Touron

[Learn more](https://2023.ethseoul.org/)

</div><!-- end column-->
<div class="column">

### BUIDL Asia

- **Proof of Contribution in Gno.land**
- Seoul, June 6, 2023
- Manfred Touron

[Learn more](https://www.buidl.asia/)

</div><!-- end column-->
<div class="column">

### Game Developer Conference

- **Side Event: Web3 Gaming Apps Powered by Gno**
- San Francisco, Mach 23, 2023
- Jae Kwon

[Watch the talk](https://www.youtube.com/watch?v=IJ0xel8lr4c)

</div><!-- end column-->
<div class="column">

### EthDenver

- **Side Event: Discover Gno.land**
- Denver, Feb 24 - Mar 5, 2023
- Jae Kwon

[Watch the talk](https://www.youtube.com/watch?v=IJ0xel8lr4c)

</div><!-- end column-->
<div class="column">

### Istanbul Blockchain Week

- Istanbul, Nov 14 - 17, 2022
- Manfred Touron

[Watch the talk](https://www.youtube.com/watch?v=JX0gdWT0Cg4)

</div><!-- end column-->
<div class="column">

### Web Summit Buckle Up and Build with Cosmos

- Lisbon, Nov 1 - 4, 2022
- Manfred Touron

</div><!-- end column-->
<div class="column">

### Cosmoverse

- Medallin, Sept 26 - 28, 2022
- Manfred Touron

[Watch the talk](https://www.youtube.com/watch?v=6s1zG7hgxMk)

</div><!-- end column-->
<div class="column">

### Berlin Blockchain Week Buckle Up and Build with Cosmos

- Berlin, Sept 11 - 18, 2022

[Watch the talk](https://www.youtube.com/watch?v=hCLErPgnavI)

</div><!-- end column-->
</div><!-- end columns-3-->`
	)
	_ = b.NewPost("", path, title, body, "2022-05-20T13:17:24Z", nil, nil)
}
