# Tendermint Classic

The design goals for Tendermint Classic (and the SDK and related libraries) are:

 * Simplicity and Legibility.
 * Parallel performance, namely ability to utilize multicore architecture.
 * Ability to evolve the codebase bug-free.
 * Debuggability.
 * Complete correctness that considers all edge cases, esp in concurrency.
 * Future-proof modular architecture, message protocol, APIs, and encapsulation
remarkably and instructively typical.
 * To minimize dependencies to evolving, complex external projects, like protobuf.
 * To become free from the influence of state actors, and mega-corpoartions like Google.
 * To be uncompromisingly opinionated, without apology.
 * To become complete, as a reference standard worthy of promoting for educational purposes.

We start with Tendermint and the Cosmos-SDK versions for cosmoshub-3, and
continue to improve the legibility of the codebase by leveraging Amino.
In the near future, Amino will become the basis for a fork of Go.

Amino Classi, I mean Tendermint Classic ... burrb... has an attitude, and if
Rick and Morty can do it, so can I, so fuck it, deal with it.

```
                                                                                                                                                                                                        
                                                                                                                                                                                                        
                                                                                                  ``                                                                                                    
                                                                                                  //:.                                                                                                  
                                                                                                 `+.-:/.                                                                                                
                                                                                                 /-...-//`              `-                                                                              
                                                                               `.`              `o......:+`           `-/o`                                                                             
                                                                               `+::-.`          /:.......:o`        `-/:.o`                                                                             
                                                                                .+.-:/:-.`     `o.........:+     `-:/-...o`                                                                             
                                                                                 /:...--:/:-.` /:..........+- `-:/:-.....o`                                                                             
                                                                                 `o.......-::/:/...........-o::--........s                                                                              
                                                                                  -+..........-......-::::::::--........-o                                                                              
                                                                                   o-............-://::-----:::///:-..../:    ``..                                                                      
                                                                                   .o.........-:/:-...............::+:-.o...-:/:/:                                                                      
                                                                                    /:.......:/:.....................:+:+:::-...o`                                                                      
                                                                                    `+.....-//........................./+......:+                                                                       
                                                                            ```````..::...-+-..-/:-.....................//....:o`                                                                       
                                                                      `.--:::::::::--.....+-...:o+//-....................s...:+`                                                                        
                                                                      `//-...............+-......:+o+++:.................o--+:`                                                                         
                                                                        `:/-............:/.....:+//:/++/++/-............-+ss.                                                                           
                                                                          `/+-..........+....-o/......-/++////////////////:+:                       -:.`                                                
                                                                            `/+-.......-+....y/////++:-..-h+++o+++++++/+yho/`                       s./+`                                               
                                                                              `//......o-...//     -:.:://s:.o/..........y:-//`                     o...o.                                              
                                                                                `/-....s....:+           `s::s:::::+///+:-h.../:                    :/...s`                                             
                                                                            `-:::.....:+.....+/          :o.-+     :    ./d..::-                    `o...-o                                             
                                                                         `://-........+-.....-/+/:-...:/+/..-s           /o//.                       o....s`                                            
                                                                      .-::-.......-/::+-.......:////::-.....:h+.       `oy:`                         o....+:                                            
                                                                     `::::/:-....//.........................+/./+//::/+y-`                           o....:+                                            
                                                                        ````.-:/.o..........................s-...:///:+-                        `.--.+....-+                                            
                                                                            `/:-.o.......:-............:...-s.........s`                   -////s:--:o:...-+`                                           
                                                                          `-/-...-+:...:/-............./+..s-........-+                   //....-o-...+...-://`                                         
                                                                          :+///::o:-//:o-.--:::::::--...:++:.........o.                   s.......:....-..-..-o                                         
                                                                          `     .+...-o-+........---::/::::-.........+                    o.-.................s`                                        
                                                                               `o-...-y-.-..............---://:::-...+                    :o..................s.                                        
                                                                               /+::::--o.......................-----.s`                   `s..................s`                                        
                                                                         ``````:..``   /+.........................-:/+                    -o.................-o                                         
                                                                      .-::------:::--.-+:+:.....................::--.`                    -+................-o`                                         
                                                                    .::.           `-+o...:+:-...............-:/.`                        .s...............:o.                                          
                                                                  `::`           `-/:o/.....-://+:--------:/:-.                           -y:............-+/`                                           
                                                                `:/.           `-+o--:s-......./s/:----...`                              -s-o/.........-/+-                                             
                                                              `-/-            -::+----:o+:--:/+:+//-/-`                                 -/o--/++::-::://-`                                              
                                                             ./:            ./-.+--------://:---/./- -/-`                             `:: o-----://yo+`                                                 
                                                           `:/`            :/` o-----------------+ /-  -+.                           `/-  `+:-----o:.s`                                                 
                                                         `-+-            `/-  /:-----------------o  +.  `+:`                        `/.     :+:--+-.-o.                                                 
                                                        `+/            `.+`  `o------------------+.  o.   :+`                      .+`        .:+///s/                                                  
                                                      `:+.           ``.+`   +:------------------:/  `o`   .+-                    -+              `/-                                                   
                                                     .+-           `` -+    `s--------------------o   `o     //`                `:/              -/`                                                    
                                                   `//           `.` -+     :/--------------------o    .o     -o.              `/:              /:                                                      
                                                 `-/`           -.  ./      o---------------------s     :/      +/`           `+-             `+.                                                       
                                                ./.           .:`  `/       s---------------------o`     +-      -+.         `+.             :+`                                                        
                                              `/:           `/-    --      -+---------------------o.      o`       /:       `o`             +:                                                          
                                            `:/            /+       +.     +:---------------------+-      `+        .:`    `+`            .o`                                                           
                                           -/`           -/+.       `o     s----------------------o/      -/          -.  `:             //`                                                            
                                         .:.           `/-`+         .+    s---------------------+/+     /:            `.`-            `+-                                                              
                                       .:.            -:` `+        .:-   `s--------------------:+-+   `s-..             .            -/`                                                               
                                      -:            .:`   -:      -/-     .o--------------------s--o  ./:: `-.                       ::`                                                                
                                     .:         `  -.`    /`     /.       -+-------------------+/..o  --`+  `.:.                   ./.                                                                  
                                     /`         `--`      o      /`       //-------------------s...s   .:+    `::`                -:`                                                                   
                                     ./          `-`      o      `+`      +:------------------o/...s    `o.     ./:             `/-`                                                                    
                                      :-           :.     o       .+      o:------------------s....s     `:.     `-/-          ./.                                                                      
                                      `/-           :.   `+        -/     o------------------o:....s      `/       `-/.       :/`                                                                       
                                       `/-           :-  .+         /:    s-----------------:s.....s      +.         `-/-` `./-                                                                         
                                        `::           :- -/          +.   s-----------------o:.....s     /s`           `.-:-.`                                                                          
                                          -/           -:::          `o`  s-----------------s......s`   -/:.                                                                                            
                                           -/`          -s-           `o  y----------------o:......s`  .+ .:                                                                                            
                                            .+`          -/            .+`y----------------s.......o. `o` `+                                                                                            
                                             `+.          ./`           :/s---------------+/.......+- +.   +                                                                                            
                                              `/-          `/`           +o---------------s........+::-    +                                                                                            
                                                :/          `:`          :+--------------/+......../+:     /`                                                                                           
                                                 -/`          :.         //--------------o.........:s      -:                                                                                           
                                                  .+`          -.        o---------------+.........:+      `+                                                                                           
                                                   `+.          /..      s--------------+-.........-o       o                                                                                           
                                                    `/:       .+/s`      y--------------+...........s       o                                                                                           
                                                      :/    `//.//      `s-------------o-...........s       +`                                                                                          
                                                       -+  -+-.//      ..s------------+/............s       /.                                                                                          
                                                        `o:.-:+-       `ss++++oooossyhs.............s       -:                                                                                          
                                                         +.`-:          :+----mmmmmmmmy.............s`      `+                                                                                          
                                                         +-`             s----Nmmmmmmmy.............s`       o                                                                                          
                                                         +.              /dhhhdddddddho.............+-       o                                                                                          
                                                         o`              /dyyyyyyyyyyyy.............//       o`                                                                                         
                                                         o`              yhyyyyyyyyyyyd.............-o       /-                                                                                         
                                                         s              +dyyyyyyyyyyyyd-.............s       -/                                                                                         
                                                         s            `shhyyyyyyyyyyyyhs.............y       `o                                                                                         
                                                         s          `-odyyyyyyyyyyyyyyyh+............s`       s                                                                                         
                                                        `s        ````ohyyyyyyyyyyyyyyyyh:.........../:       o`                                                                                        
                                                        `s            hhyyyyyyyyyyyyyyyyhy...........-o       +.                                                                                        
                                                        .o           -dyyyyyyyyhhhhyyyyyyho...........s       /-                                                                                        
                                                        -/           s:-------/hhhyyyyyyyyd:..........o.      ./                                                                                        
                                                        /:          `m/........+hyyyyyyyyyhh..........:+      `o                                                                                        
                                                        o.          +dy.........shyyyyyyyyyho..........s       s                                                                                        
                                                        s          `dhd-........-yhyyyyyyyyyd:.........o-      o`                                                                                       
                                                        s          /dyh+........./hyyyyyyyyyhh-........-o      +.                                                                                       
                                                       `o         `dhyhy..........ohyyyyyyyyyhs.........s`     ::                                                                                       
                                                       :/         ohyyyd-..........shyyyyyyyyyd/........:+     `+                                                                                       
                                                       +.        -dyyyyho..........-hhyyyyyyyyhh-........s`     o                                                                                       
                                                       s        `yhyyyyyh...........:hyyyyyyyyyhs........:/     o                                                                                       
                                                      `o        +hyyyyyyd:...........+hyyyyyyyyyd+........s`    +`                                                                                      
                                                      -/       -dyyyyyyyhs............shyyyyyyyyyd-.......:+    :.                                                                                      
                                                      +.      .hhyyyyyyyyd.............yhyyyyyyyyhh........o`   .:                                                                                      
                                                      o      `sdyyyyyyyyyd/............-dhyyyyyyyyho.......-+   `+                                                                                      
                                                     `+      o-hyyyyyyyyyhh.............+dyyyyyyyyyh:......./-   +                                                                                      
                                                     /.     o:.ohyyyyyyyyyd-.............shyyyyyyyyyh-.......+   /                                                                                      
                                                     +     +:..:dyyyyyyyyyhs..............yhyyyyyyyyhy.......-/  /`                                                                                     
                                                    `/    +:....hyyyyyyyyyhd..............-hhyyyyyyyyho.......:. .-                                                                                     
                                                    :.   /:.....shyyyyyyyyym:..............-hhyyyyyyyyd:......./ `:                                                                                     
                                                    /   /-....../hyyyyyyyyydy///////////////odyyyyyyyyhh-.......: /                                                                                     
                                                   `: `/-.......-dyyyyyyyyyhs````````````````/hyyyyyyyyhh/////:--.:                                                                                     
                                                   :``/-...-::///hyyyyyyyyyyy                `yhyyyyyyyyd.````.-://`                                                                                    
                                                  `:./:::::-```` yyyyyyyyyyyh`                /hyyyyyyyyh+       `:.                                                                                    
                                                  .//:.``        syyyyyyyyyyd`                .dyyyyyyyyhs        ``                                                                                    
                                                  .`             syyyyyyyyyyd`                `hyyyyyyyyhy                                                                                              
                                                                 syyyyyyyyyyd`                 yhyyyyyyyhh                                                                                              
                                                                 shyyyyyyyyyd`                 ohyyyyyyyyd                                                                                              
                                                                 shyyyyyyyyyd`                 ohyyyyyyyyd                                                                                              
                                                                 shyyyyyyyyyh`                 shyyyyyyyhh                                                                                              
                                                                 yhyyyyyyyyyy                  yyyyyyyyyhs                                                                                              
                                                                 yyyyyyyyyyys                  yyyyyyyyyh+                                                                                              
                                                                 hyyyyyyyyyyh`                 hyyyyyyyyd:                                                                                              
                                                                `dyyyyyyyyyyh                 `dyyyyyyyyd.                                                                                              
                                                                `dyyyyyyyyyhy                 `dyyyyyyyyd`                                                                                              
                                                                .dyyyyyyyyyho                 .dyyyyyyyyh`                                                                                              
                                                                -dyyyyyyyyyh/                 -dyyyyyyyhy                                                                                               
                                                                :dyyyyyyyyyd-                 :dyyyyyyyho                                                                                               
                                                                /hyyyyyyyyyd.                 +hyyyyyyyd/                                                                                               
                                                                +hyyyyyyyyyd`                 ohyyyyyyyd-                                                                                               
                                                                shyyyyyyyyhh                  shyyyyyyyd`                                                                                               
                                                                yhyyyyyyyyhs                  yhyyyyyyyd                                                                                                
                                                                hyyyyyyyyyh+                  dyyyyyyyhy                                                                                                
                                                                dyyyyyyyyyd:                 `dyyyyyyyho                                                                                                
                                                               `dyyyyyyyyyd.                 `myyyyyyyh/                                                                                                
                                                               `dyyyyyyyyyd`                 .dyyyyyyyd.                                                                                                
                                                               -dyyyyyyyyyh                  -dyyyyyyyd`                                                                                                
                                                               :hyyyyyyyyhs                  /hyyyyyyyy                                                                                                 
                                                               +hyyyyyyyyh+                  +hyyyyyyys                                                                                                 
                                                               ohyyyyyyyyh:                  oyyyyyyyh/                                                                                                 
                                                               ohhhhhhhhyy`                  /hhhhyyyh-                                                                                                 
                                                               `/:----..-/                   `o.-/+osy`                                                                                                 
                                                                :-      ./                    o     `/                                                                                                  
                                                               .//---:::++`                   +.--:/ohs:`                                                                                               
                                                              `hdddddddddds.                 `hhddddddddh+.                                                                                             
                                                              oddddddddddddd/`               oddddddddddddds:`                                                                                          
                                                             .dddddddddddddddy-             .mdddddddddddddddy/.                                                                                        
                                                             sddddddddddddddddd+`           omdddddddddddddddddho-`                                                                                     
                                                            .mddddddddddddddddddy-`         yddddddddddddddddddddds-`                                                                                   
                                                            smddddddddddddddddddddo.        dddddddddddddddddddddddds-`                                                                                 
                                                            /++++++++++++++++++++++:       `rippedfromkineticsqurrel/`                                                                                  
                                                                                                                                                                                                        
```

But is it really?
