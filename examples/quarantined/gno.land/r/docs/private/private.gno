package private

import "strings"

// Documentation for private realms in Gno.
func Render(path string) string {
	output := `# Private Realms

## Introduction

This document describes the ±private± flag in ±gnomod.toml±.  
This flag is **optional** and can only be used by **realms**.

> [!NOTE]
> The ±private± flag does not make a realm secret or hidden.  
> All code remains open-source and visible on-chain.

## Declaring a Private Realm

To mark a realm as private, add the following ±private± value in your ±gnomod.toml± file:

±±±toml
module = "gno.land/r/demo/privaterealm"
gno = "0.9"
private = true
±±±

## Rules of a Private Realm

A realm marked as private follows strict rules:

- It cannot be imported by other packages.
- It can be re-uploaded and the new version fully overwrites the old one.
- Memory, pointers, or types defined in this package cannot be stored elsewhere.

These restrictions aim to achieve one main goal:

> [!TIP]
> **Private realms are swappable.**  
> You can redeploy or update them without impacting other realms.  

## Use Cases

Private realms are useful when:

- The realm represents a **home realm** for a specific user (±r/username/home±).
- The realm is **designed to evolve** and may need to be replaced later.
- You want to ensure that no external state depends on its internal types or values.
`
	output = strings.ReplaceAll(output, "±", "`")
	return output
}
