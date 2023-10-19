package markdown

// this package can be used to test markdown rendering engines.

func Render(path string) string {
	output := `_imported from https://github.com/markedjs/marked/blob/master/docs/demo/quickref.md_

Markdown Quick Reference
========================

This guide is a very brief overview, with examples, of the syntax that [Markdown] supports. It is itself written in Markdown and you can copy the samples over to the left-hand pane for experimentation. It's shown as *text* and not *rendered HTML*.

[Markdown]: http://daringfireball.net/projects/markdown/


Simple Text Formatting
======================

First thing is first. You can use *stars* or _underscores_ for italics. **Double stars** and __double underscores__ for bold. ***Three together*** for ___both___.

Paragraphs are pretty easy too. Just have a blank line between chunks of text.

> This chunk of text is in a block quote. Its multiple lines will all be
> indented a bit from the rest of the text.
>
> > Multiple levels of block quotes also work.

Sometimes you want to include code, such as when you are explaining how ` + "`<h1>`" + ` HTML tags work, or maybe you are a programmer and you are discussing ` + "`someMethod()`" + `.

If you want to include code and have new
lines preserved, indent the line with a tab
or at least four spaces:

    Extra spaces work here too.
    This is also called preformatted text and it is useful for showing examples.
    The text will stay as text, so any *markdown* or <u>HTML</u> you add will
    not show up formatted. This way you can show markdown examples in a
    markdown document.

>     You can also use preformatted text with your blockquotes
>     as long as you add at least five spaces.


Headings
========

There are a couple of ways to make headings. Using three or more equals signs on a line under a heading makes it into an "h1" style. Three or more hyphens under a line makes it "h2" (slightly smaller). You can also use multiple pound symbols (` + "`#`" + `) before and after a heading. Pounds after the title are ignored. Here are some examples:

This is H1
==========

This is H2
----------

# This is H1
## This is H2
### This is H3 with some extra pounds ###
#### You get the idea ####
##### I don't need extra pounds at the end
###### H6 is the max


Links
=====

Let's link to a few sites. First, let's use the bare URL, like <https://www.github.com>. Great for text, but ugly for HTML.
Next is an inline link to [Google](https://www.google.com). A little nicer.
This is a reference-style link to [Wikipedia] [1].
Lastly, here's a pretty link to [Yahoo]. The reference-style and pretty links both automatically use the links defined below, but they could be defined *anywhere* in the markdown and are removed from the HTML. The names are also case insensitive, so you can use [YaHoO] and have it link properly.

[1]: https://www.wikipedia.org
[Yahoo]: https://www.yahoo.com

Title attributes may be added to links by adding text after a link.
This is the [inline link](https://www.bing.com "Bing") with a "Bing" title.
You can also go to [W3C] [2] and maybe visit a [friend].

[2]: https://w3c.org (The W3C puts out specs for web-based things)
[Friend]: https://facebook.com "Facebook!"

Email addresses in plain text are not linked: test@example.com.
Email addresses wrapped in angle brackets are linked: <test@example.com>.
They are also obfuscated so that email harvesting spam robots hopefully won't get them.


Lists
=====

* This is a bulleted list
* Great for shopping lists
- You can also use hyphens
+ Or plus symbols

The above is an "unordered" list. Now, on for a bit of order.

1. Numbered lists are also easy
2. Just start with a number
3738762. However, the actual number doesn't matter when converted to HTML.
1. This will still show up as 4.

You might want a few advanced lists:

- This top-level list is wrapped in paragraph tags
- This generates an extra space between each top-level item.

- You do it by adding a blank line

- This nested list also has blank lines between the list items.

- How to create nested lists
  1. Start your regular list
  2. Indent nested lists with two spaces
  3. Further nesting means you should indent with two more spaces
    * This line is indented with four spaces.

- List items can be quite lengthy. You can keep typing and either continue
them on the next line with no indentation.

- Alternately, if that looks ugly, you can also
  indent the next line a bit for a prettier look.

- You can put large blocks of text in your list by just indenting with two spaces.

  This is formatted the same as code, but you can inspect the HTML
  and find that it's just wrapped in a ` + "`<p>`" + ` tag and *won't* be shown
  as preformatted text.

  You can keep adding more and more paragraphs to a single
  list item by adding the traditional blank line and then keep
  on indenting the paragraphs with two spaces.

  You really only need to indent the first line,
but that looks ugly.

- Lists support blockquotes

  > Just like this example here. By the way, you can
  > nest lists inside blockquotes!
  > - Fantastic!

- Lists support preformatted text

      You just need to indent an additional four spaces.

Even More
=========

Horizontal Rule
---------------

If you need a horizontal rule you just need to put at least three hyphens, asterisks, or underscores on a line by themselves. You can also even put spaces between the characters.

---
****************************
_ _ _ _ _ _ _

Those three all produced horizontal lines. Keep in mind that three hyphens under any text turns that text into a heading, so add a blank like if you use hyphens.

Images
------

Images work exactly like links, but they have exclamation points in front. They work with references and titles too.

![Google Logo](https://www.google.com/images/errors/logo_sm.gif) and ![Happy].

[Happy]: https://wpclipart.com/smiley/happy/simple_colors/smiley_face_simple_green_small.png ("Smiley face")

Inline HTML
-----------

If markdown is too limiting, you can just insert your own <strike>crazy</strike> HTML. Span-level HTML <u>can *still* use markdown</u>. Block level elements must be separated from text by a blank line and must not have any spaces before the opening and closing HTML.

<div style='font-family: "Comic Sans MS", "Comic Sans", cursive;'>
It is a pity, but markdown does **not** work in here for most markdown parsers.
[Marked] handles it pretty well.
</div>

Components
==========

Jumbotron component
-------------------

A Jumbotron is a layout component used to showcase key messages.


:::jumbotron

## Jumbotron title

This is a Jumbotron component you can fill with regular markdown.

:::/jumbotron

Stack component
---------------

Stack is a layout component used to group elements together and apply a space between them. For example, it is usefull with grouped links.

:::stack
[Stack link 1](/)
[Stack link 2](/)
:::/stack

Box component
-------------

Box is the most abstract component that renders a div element. It is used as a group component.

:::box
**This content is in a box**
:::/box


Columns component
-----------------

Columns is a layout for managing columns layouts. It takes the numb of cols you want (*from 1 to 6 columns*) to create as argument and need to be filled with as much content as you mentioned in this argument. You can manage multiple element by grouping them in box components. 

*Nested columns are not currently allowed*

:::columns(2)
:::box

### Col 1

Content 1

:::/box
:::box

### Col 2

Content 2

:::/box
:::/columns

:::columns(3)
:::box

### Col 1

Content 1

:::/box
:::box

### Col 2

Content 2

:::/box
:::box

### Col 3

Content 3

:::/box
:::/columns

:::columns(4)
:::box

### Col 1

Content 1

:::/box
:::box

### Col 2

Content 2

:::/box
:::box

### Col 3

Content 3

:::/box
:::box

### Col 4

Content 4

:::/box
:::/columns

:::columns(5)
:::box

### Col 1

Content 1

:::/box
:::box

### Col 2

Content 2

:::/box
:::box

### Col 3

Content 3

:::/box
:::box

### Col 4

Content 4

:::/box
:::box

### Col 5

Content 5

:::/box
:::/columns

:::columns(6)
:::box

### Col 1

Content 1

:::/box
:::box

### Col 2

Content 2

:::/box
:::box

### Col 3

Content 3

:::/box
:::box

### Col 4

Content 4

:::/box
:::box

### Col 5

Content 5

:::/box
:::box

### Col 6

Content 6

:::/box
:::/columns

Button component
----------------

Button component is used as CTA (for a link) or to display a button element that you will be able to use to trigger an action or event through Javascript.
Button is single line component that takes few arguments: the content value as first argument and an optional url as second argument. 

_Note that arguments are separated by comas and you can use quotes to include comas within argument._

:::button("Link button with comma, needs quotes", https://gno.land)/

:::button(Plain button)/

Accordion component
-------------------

Accordions display a list of high-level options that can expand/collapse to reveal more information. It takes the button text as argument.

*Todo: add argument to set the state at first (open or closed)*

:::accordion(Accordion open/close button)
Accordion content
:::/accordion

Tabs component
--------------

A component to create accessible tabs. It takes the button text as arguments.

:::tabs(1st tab button text, 2nd tab button text)
:::box

## 1st Tab Title

1st tab Content
:::/box

:::box

## 2nd Tab Title

2nd tab Content
:::/box
:::/tabs


Alert component
---------------

Alerts are used to communicate a state that affects a system, feature or page. The component take the alert style as argument: "warning", "info", "danger", "success".

:::alert(warning)
Warning content
:::/alert

:::alert(info)
Info content
:::/alert

:::alert(danger)
Danger content
:::/alert

:::alert(success)
Success content
:::/alert

Breadcrumb component
--------------------

Breadcrumbs is a navigation pattern that helps users understand the hierarchy of a website. It requires to be filled by an ordered list.

:::breadcrumb

1. [Home](https://gno.land)
2. [Foo](https://gno.land)
3. [Bar](https://gno.land)

:::/breadcrumb

Dropdown component
------------------

Dropdowns are toggleable, contextual overlays for displaying lists of links and more. It requires to be filled by a list and takes the dropdown text button as argument.

:::dropdown(Dropdown - click here)

- [Home](https://gno.land)
- [Foo](https://gno.land)
- [Bar](https://gno.land)

:::/dropdown

Pagination component
--------------------

Breadcrumbs is a navigation component that helps users navigate between website pages. It requires to be filled by an ordered list and can take the description of the navigation (for a11y purpose) as argument.

:::pagination(Article pages)

- [1](/)
- [2](/)
- [3](/)

:::/pagination

Form component
--------------

This form component help you create forms. It has to be filled with dedicated components listed bellow. It takes several argument that are used to customize the form settings: 

- First argument is the Action (eg. link: "https://gno.land") 
- Second argument is the Method  (eg. "post", "get", ...)

:::form(/, get)

### Form inputs components

Input component is a component that is used to get user input in a text field. It takes 2 arguments.

- First argument is the input type (eg. "text", "date", "password", "number"...)
- Second argument (optional) is the label text content
- Third argument (optional) is the placeholer
- Fourth argument (optional) is the value content

You can insert text inside the component and it will be used as placeholder content. If not, you can use the shorthand component declaration by adding a "/" direclty after the arguments (see date input component declaration bellow).

:::input(text, Label, Text)/

:::input(number, Label, 10, 3)/

:::input(date, Label)/

:::input(password, Password)/


### Form textarea component

The Textarea component allows you to easily create multi-line text inputs. The arguments are the optional label text content and placeholder.

:::textarea(Label, Placeholder)/

### Form check components

Checkbox component is used in forms when a user needs to select multiple values from several options. By using the first parameter, you can choose whether you want a "radio" or a "checkbox" list. The second argument (optional) is the label text content.

*Todo: add a way to preselect options*

:::check(radio, Radio Label)

- Radio 1
- Radio 2
- Radio 3

:::/check

:::check(checkbox)

- Checkbox 1
- Checkbox 2
- Checkbox 3

:::/check

### Form select component

Similar to the previous one, select component is a component that allows users pick a value from predefined options. The argument (optional) is the label text content.

:::select(Select Label)

- Select 1
- Select 2
- Select 3

:::/select

### Form button component

This component is a button component dedicated to form control. Depending on the argument, it can be used to submit, reset... a form. The component takes the button type as firt argument and the button value as second argument.

:::action("reset", "Reset")/

:::action("submit", "Submit")/

:::/form
`
	return output
}
