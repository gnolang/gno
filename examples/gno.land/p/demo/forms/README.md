# Gno forms

gno-forms is a package which demonstrates a form editing and sharing application in gno

## Features
- **Form Creation**: Create new forms with specified titles, descriptions, and fields.
- **Form Submission**: Submit answers to forms.
- **Form Retrieval**: Retrieve existing forms and their submissions.
- **Form Deadline**: Set a precise time range during which a form can be interacted with.

## Field Types
The system supports the following field types:

type|example
-|-
string|`{"label": "Name", "fieldType": "string", "required": true}`
number|`{"label": "Age", "fieldType": "number", "required": true}`
boolean|`{"label": "Is Student?", "fieldType": "boolean", "required": false}`
choice|`{"label": "Favorite Food", "fieldType": "['Pizza', 'Schnitzel', 'Burger']", "required": true}`
multi-choice|`{"label": "Hobbies", "fieldType": "{'Reading', 'Swimming', 'Gaming'}", "required": false}`

## Web-app

The external repo where the initial development took place and where you can find the frontend is [here](https://github.com/agherasie/gno-forms). 