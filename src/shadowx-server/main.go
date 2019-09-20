package main

import (
	"flag"
	"shadowx-server/server"

	"github.com/ije/rex"
)

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Hello World</title>
	<style>
		body {
			margin: 120px auto;
			max-width: 600px;
		}
		h1 {
			font-size: 30px;
			font-weight: 500;
			color: #111;
		}
		p {
			font-size: 14px;
			color: #333;
		}
	</style>
</head>
<body>
	<h1>Ipsum dolor</h1>
	<p>Consequat duis autem vel eum iriure dolor in. Ad minim veniam quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea. Legentis in iis qui facit eorum claritatem Investigationes. Consuetudium lectorum mirum est notare quam littera gothica quam nunc putamus parum claram anteposuerit litterarum formas.</p>
	<p>Littera gothica quam nunc putamus parum claram anteposuerit litterarum formas humanitatis per seacula quarta. Insitam est usus legentis in iis qui facit eorum claritatem. Tincidunt ut laoreet dolore magna aliquam erat volutpat. Molestie consequat vel illum dolore eu feugiat nulla. Velit esse facilisis at vero eros et accumsan et iusto odio dignissim? Placerat facer possim assum typi non habent claritatem Investigationes demonstraverunt lectores legere me.</p>
	<p>Enim ad minim veniam quis nostrud exerci tation? Dolore te feugait nulla facilisi nam liber tempor cum soluta nobis. Formas humanitatis per seacula quarta decima et quinta decima eodem modo. Claritatem insitam est usus legentis in iis qui facit eorum claritatem? Dolor sit amet consectetuer adipiscing elit sed diam nonummy nibh euismod. Illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto. Delenit augue duis eleifend option congue nihil imperdiet doming id quod mazim placerat facer possim. Parum claram anteposuerit litterarum typi qui nunc nobis videntur parum clari, fiant sollemnes in.</p>
	<p>Adipiscing elit sed diam nonummy, nibh euismod tincidunt. Hendrerit in vulputate velit: esse molestie consequat vel illum dolore eu feugiat nulla facilisis at vero? Dolor sit amet consectetuer ut laoreet dolore magna aliquam.</p>
	<p>Laoreet dolore magna aliquam erat volutpat ut wisi enim ad minim veniam quis nostrud exerci. Qui sequitur mutationem consuetudium lectorum mirum est, notare quam littera gothica! Per seacula quarta decima et quinta decima eodem. Soluta nobis eleifend option congue nihil imperdiet doming id. Legunt saepius claritas est etiam processus dynamicus quam. Ullamcorper suscipit lobortis nisl ut, aliquip ex ea commodo. Claram anteposuerit litterarum formas humanitatis modo typi qui nunc nobis videntur parum clari fiant sollemnes.</p>
	<p>Etiam processus dynamicus, qui sequitur mutationem consuetudium lectorum. Consectetuer adipiscing elit sed diam nonummy nibh euismod tincidunt ut laoreet dolore?</p>
	<p>Non habent claritatem insitam est usus legentis in iis qui. Qui blandit praesent luptatum zzril delenit augue duis dolore te feugait? Nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat ut wisi enim ad! Litterarum formas humanitatis per seacula quarta decima et quinta. Accumsan et iusto odio dignissim nulla facilisi nam liber. Placerat facer possim assum typi facit eorum claritatem Investigationes demonstraverunt lectores legere me lius!</p>
	<p>Ipsum dolor sit amet consectetuer adipiscing elit sed diam nonummy nibh euismod. Quod ii legunt saepius claritas est etiam processus dynamicus, qui sequitur mutationem consuetudium lectorum mirum est. Eleifend option congue nihil imperdiet doming id quod; mazim placerat facer possim! Feugait nulla facilisi nam liber tempor cum soluta nobis assum typi non habent claritatem insitam est.</p>
	<p>Praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi nam. Iusto odio dignissim qui blandit liber tempor cum soluta nobis eleifend option.</p>
	<p>Est usus legentis in, iis qui facit eorum. Esse molestie consequat vel illum dolore eu feugiat. Veniam quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex; ea commodo consequat duis. Vel eum iriure dolor in hendrerit in vulputate velit nulla?</p>
	<p>Congue nihil imperdiet doming id quod mazim placerat facer possim. Habent claritatem insitam est usus legentis in iis.</p>
	<p>Facilisi nam liber tempor cum soluta nobis eleifend option congue nihil. Mutationem consuetudium lectorum mirum est notare quam littera.</p>
</body>
</html>
`

func main() {
	d := flag.Bool("d", false, "debug mode")
	flag.Parse()

	rex.Get("/", func(ctx *rex.Context) {
		ctx.HTML(indexHTML)
	})

	rex.Get("/api/ws", func(ctx *rex.Context) {
		server.Serve(ctx.W, ctx.R, *d)
	})

	if *d {
		rex.Serve(rex.Config{
			Port: 80,
		})
		return
	}

	rex.Serve(rex.Config{
		TLS: rex.TLSConfig{
			Port: 443,
			AutoTLS: rex.AutoTLSConfig{
				AcceptTOS: true,
				CacheDir:  "/etc/shadowx/cert-cache",
			},
		},
	})
}
