package home

import (
	"std"

	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"
)

var (
	admin   = std.Address("g1qdwemgqgqm7s42gy6xcj7f22uhfdzr82st3sy3")
	project = md.Paragraph("Discover the first Optimistic Oracle on Gnoland.") +
		md.Paragraph("This project is a proof of concept for an oracle system that leverages community participation and token incentives to provide reliable off-chain data to smart contracts. It is inspired by UMA's Optimistic Oracle but adapted for the Gno ecosystem.") +
		md.Paragraph("More about the project: "+md.Link("Here", "gno.land/r/intermarch3/oo$source"))
	title             = "The home realm of Intermarch3."
	githubUsername    = "intermarch3"
	pocInnovationName = "PoCInnovation"
	art               = `
                    +=         +=                   
                  @@@@@@@@@@@@@@@@*                 
                 @@@@@@@@@@@@@@@@@@%                
                @@@@@@@@@@@@@@@@@@@@+               
               :@@@@@@@@@@@@@@@@@@@@@               
               @@@@@@@@@@@@@@@@@@@@@@-              
          ....+@@@@@@@@@@@@@@@@@@@@@@%:...          
        :@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@-        
                                                    
                                                    
                  %@@@:      =@@@#                  
                 %@@@@@*    #@@@@@*                 
                 =@@@@@      @@@@@:                 
                    '          '                    `
)

func init() {
	hor.Register(cross, "Intermarch3 Home Realm", title)
}

func renderProject() string {
	out := md.H2("⚒ Project highlight")
	out += md.Paragraph(project)
	return out
}

func renderArt() string {
	out := md.H2("🎨 Art")
	out += md.CodeBlock(art)
	return out
}

func renderBody() string {
	return md.Columns([]string{
		renderProject(),
		renderArt(),
	}, false)
}

func renderTitle() string {
	return md.H1(title)
}

func renderFooter() string {
	out := md.HorizontalRule()
	out += md.BulletList([]string{
		md.Link("Home", "home"),
		md.Link("GitHub: @"+githubUsername, "https://github.com/"+githubUsername),
		md.Link("PoC-Innovation", "https://github.com/"+pocInnovationName),
	})
	return out
}

func Render(_ string) string {
	out := renderTitle()
	out += md.Paragraph("Welcome to the home realm of Intermarch3, a developer and passionate about blockchain.") + "\n\n"
	out += renderBody()
	out += renderFooter()

	return out
}

// -- Update functions --

func UpdateProject(_ realm, newProject string) {
	if std.PreviousRealm().Address() != admin {
		panic("only admin can update the project description")
	}
	project = newProject
}

func UpdateArt(_ realm, newArt string) {
	if std.PreviousRealm().Address() != admin {
		panic("only admin can update the art")
	}
	art = newArt
}
