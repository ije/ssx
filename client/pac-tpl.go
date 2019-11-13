package client

import "text/template"

const pacTS = `
var hasOwnProperty = Object.hasOwnProperty;
var proxy = '{{ .proxy }}; DIRECT;';
var direct = 'DIRECT;';
var hosts = {
    // gfwlist{{range $host := .gfwList}}
    '{{$host}}': 1,{{end}}
    {{with .custom}}// custom{{range $host := .}}
    '{{$host}}': 1,{{end}}{{end}}
};

function FindProxyForURL(url, host) {
    var suffix;
    var pos = host.lastIndexOf('.');    
    while(1) {
        pos = host.lastIndexOf('.', pos - 1);
        if (pos <= 0) {
            if (hasOwnProperty.call(hosts, host)) {
                return proxy;
            } else {
                return direct;
            }
        }
        suffix = host.substring(pos + 1);
        if (hasOwnProperty.call(hosts, suffix)) {
            return proxy;
        }
    }
}
`

var pacTpl *template.Template

func init() {
	pacTpl = template.Must(template.New("pac").Parse(pacTS))
}
