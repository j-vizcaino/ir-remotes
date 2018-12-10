package ui

var (
	rawIndexTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Remotes</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bulma/0.7.2/css/bulma.min.css">
    <script defer src="https://use.fontawesome.com/releases/v5.3.1/js/all.js"></script>
	<script>
function sendCommand(remote, command) {
	var url = "/api/remotes/" + remote + "/" + command;
    var xhttp = new XMLHttpRequest();
	xhttp.open("POST", url, true);
	xhttp.send();
}
	</script>
  </head>
  <body>
  <section class="section">
  {{ range $_, $remote := . -}}
      <div class="container box">
        <h2 class="title">
            {{ $remote.Text }}
        </h2>
        <div class="buttons">
          {{ range $_, $button := $remote.Buttons -}}
		  <a class="button {{ $button.Class }}" onclick="sendCommand('{{ $remote.Name }}', '{{ $button.Name }}')">
          {{ if ne $button.Icon "" -}}
            <span class="icon {{ if ne $button.Text "" }}is-small{{ end }}">
            <i class="fas fa-{{ $button.Icon }}"></i>
            </span>
          {{ end -}}
          {{- if ne $button.Text "" }}<span>{{ $button.Text }}</span>{{ end }}
          </a>
          {{ end -}}
        </div>
      </div>
  {{ end -}}
  </section>
  </body>
</html>
`
)
