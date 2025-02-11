package home

func Render(path string) string {
   home := New()
   home.h1("Mason's Realm")
   home.p("Welcome to my realm.")
   home.im("https://i.imgflip.com/7ze8dc.jpg","Placeholder")
   home.p("Dgnonut")
   home.p("Coming SOON")
   return home.Render()
}