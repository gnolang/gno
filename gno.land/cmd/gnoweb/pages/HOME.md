# Welcome to **Gno.land**

- [About Gno.land](/about)
- [Blogs](/r/gnoland/blog)
- [Install `gnokey`](https://github.com/gnolang/gno/tree/master/gno.land/cmd/gnokey)
- [Acquire testnet tokens](/faucet)
- [Game of Realms](/game-of-realms) - An open worldwide competition for developers to build the best Gnolang smart-contracts.

## Jumbotron component

:::jumbotron

### Jumbotron Title

This is a Jumbotron component you can fill with regular markdown.

:::stack
[Stack link 1](/about)
[Stack link 1](https://github.com/gnolang)
:::stack/
:::jumbotron/

```markdown
:::jumbotron

### Title

Content

:::jumbotron/
```

## Stack component

:::stack
[Stack link 1](/about)
[Stack link 1](https://github.com/gnolang)
:::stack/

```markdown
:::stack
[Stack link 1](/about)
[Stack link 1](https://github.com/gnolang)
:::stack/
```

## columns component

:::columns (2)
:::box

### Col 1

Content 1

:::box/
:::box

### Col 2

Content 2

:::box/
:::columns/

:::columns (3)
:::box

### Col 1

Content 1

:::box/
:::box

### Col 2

Content 2

:::box/
:::box

### Col 3

Content 3

:::box/
:::columns/

:::columns (4)
:::box

### Col 1

Content 1

:::box/
:::box

### Col 2

Content 2

:::box/
:::box

### Col 3

Content 3

:::box/
:::box

### Col 4

Content 4

:::box/
:::columns/

:::columns (5)
:::box

### Col 1

Content 1

:::box/
:::box

### Col 2

Content 2

:::box/
:::box

### Col 3

Content 3

:::box/
:::box

### Col 4

Content 4

:::box/
:::box

### Col 5

Content 5

:::box/
:::columns/

:::columns (6)
:::box

### Col 1

Content 1

:::box/
:::box

### Col 2

Content 2

:::box/
:::box

### Col 3

Content 3

:::box/
:::box

### Col 4

Content 4

:::box/
:::box

### Col 5

Content 5

:::box/
:::box

### Col 6

Content 6

:::box/
:::columns/

```markdown
:::columns (2)
:::box

First column content

:::box/
:::box

First column content

:::box/
:::columns/

From 1 to 6 column grouped by box component
```

## Button component

:::button (https://gno.land)
Link button
:::button/

:::button
State button
:::button/

```markdown
:::button (https://gno.land)
Link button
:::button/

or without link to create a button element instead of a link one

:::button
State button
:::button/
```

## Accordion component

:::accordion (Accordion button)
Accordion content
:::accordion/

```markdown
:::accordion (Accordion button content)
Accordion content
:::accordion/
```

## Tabs component

:::tabs (1st tab button text)(2nd tab button text)
:::box

## 1st Tab Title

1st tab Content
:::box/

:::box

## 2nd Tab Title

2nd tab Content
:::box/
:::tabs/

```markdown
:::tabs (Tab button text 1)(Tab button text 1)
:::box
Tab Content 1
:::box/

:::box
Tab Content 2
:::box/
:::tabs/
```

## Alert component

:::alert (warning)
Warning content
:::alert/

:::alert (info)
Info content
:::alert/

:::alert (danger)
Danger content
:::alert/

:::alert (success)
Success content
:::alert/

```markdown
:::alert (success | danger | info | warning)
Alert content
:::alert/
```

## Breadcrumb component

:::breadcrumb

1. [Home](https://gno.land)
2. [Foo](https://gno.land)
3. [Bar](https://gno.land)

:::breadcrumb/

```markdown
:::breadcrumb

1. [Home](https://gno.land)
2. [Foo](https://gno.land)
3. [Bar](https://gno.land)

:::breadcrumb/
```

## Dropdown component

:::dropdown (Dropdown - click here)

1. [Home](https://gno.land)
2. [Foo](https://gno.land)
3. [Bar](https://gno.land)

:::dropdown/

```markdown
:::dropdown (Dropdown - click here)

1. [Home](https://gno.land)
2. [Foo](https://gno.land)
3. [Bar](https://gno.land)

:::dropdown/
```

## Pagination component

:::pagination (Article pages)

1. [1](https://gno.land)
2. [2](https://gno.land)
3. [3](http://127.0.0.1:8888/)

:::pagination/

```markdown
:::pagination

1. [1](https://gno.land/1)
2. [2](https://gno.land/2)
3. [3](https://gno.land/3)

:::pagination/
```

## Form component

```markdown
:::form (https://gno.land) (post | get | ...)
Form content & components such as input (see below)
:::form/

First argument is action and second one is method
```

:::form (/)(get)

### Form inputs components

:::form-input (text)(Input Label)
Placeholder
:::form-input/
:::form-input (date)(Input Label)/
:::form-input (number)
Number
:::form-input/
:::form-input (password)
Password
:::form-input/

```markdown
With placeholder content

:::form-input (text | number | password | ...)(Label?)
Password
:::form-input/

Without placeholder content

:::form-input (date)(Label?)/

Without label

:::form-input (number)
Number
:::form-input/
```

### Form textarea component

:::form-textarea (Label)
Placeholder
:::form-textarea/

```markdown
:::form-textarea (Label?)
Placeholder
:::form-textarea/

Label param is optional
```

### Form check components

:::form-check (radio)(Radio Label)

- Radio 1
- Radio 2
- Radio 3

:::form-check/

:::form-check (checkbox)

- Checkbox 1
- Checkbox 2
- Checkbox 3

:::form-check/

```markdown
:::form-check (checkbox | radio)(Label?)

- Checkbox 1
- Checkbox 2
- Checkbox 3

:::form-check/

Label param is optional
```

### Form select component

:::form-select (Select Label)

- Select 1
- Select 2
- Select 3

:::form-select/

```markdown
:::form-select (Label?)

- Select 1
- Select 2
- Select 3

:::form-select/

Label is optional
```

### Form buttons component

:::form-button (reset)
Reset
:::form-button/
:::form-button (submit)
Submit
:::form-button/

```markdown
:::form-button (submit | reset | ...)
Submit
:::form-button/
```

:::form/

# Explore new packages.

- r/gnoland
  - [/r/gnoland/blog](/r/gnoland/blog)
  - [/r/gnoland/faucet](/r/gnoland/faucet)
- r/system
  - [/r/system/names](/r/system/names)
  - [/r/system/rewards](/r/system/rewards)
  - [/r/system/validators](/r/system/validators)
- r/demo
  - [/r/demo/banktest](/r/demo/banktest)
  - [/r/demo/boards](/r/demo/boards)
  - [/r/demo/foo20](/r/demo/foo20)
  - [/r/demo/nft](/r/demo/nft)
  - [/r/demo/types](/r/demo/types)
  - [/r/demo/users](/r/demo/users)
  - [/r/demo/groups](/r/demo/groups)
- p/demo
  - [/p/demo/avl](/p/demo/avl)
  - [/p/demo/blog](/p/demo/blog)
  - [/p/demo/flow](/p/demo/flow)
  - [/p/demo/gnode](/p/demo/gnode)
  - [/p/demo/grc/exts](/p/demo/grc/exts)
  - [/p/demo/grc/grc20](/p/demo/grc/grc20)
  - [/p/demo/grc/grc721](/p/demo/grc/grc721)

# Other Testnets

- **[staging.gno.land](https://staging.gno.land) (wiped every commit to master)**
- _[test3.gno.land](https://test3.gno.land) (latest)_
- _[test2.gno.land](https://test2.gno.land) (archive)_
- _[test1.gno.land](https://test1.gno.land) (archive)_

**This is a testnet.**
Package names are not guaranteed to be available for production.

# Social

Check out our [community projects](https://github.com/gnolang/awesome-gno).

Official channel: [Discord](https://discord.gg/S8nKUqwkPn)<br />
Other channels: [Telegram](https://t.me/gnoland) [Twitter](https://twitter.com/_gnoland) [Youtube](https://www.youtube.com/@_gnoland)
