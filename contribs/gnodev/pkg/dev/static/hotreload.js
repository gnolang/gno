(function() {
    var eventsReload = {{ .ReloadEvents | jsEventsArray }};
    var ws = new WebSocket('ws://{{- .Remote -}}');

    ws.onmessage = function(event) {
        try {
            var message = JSON.parse(event.data);
            console.log('receiving message:', message);
            if (eventsReload.includes(message.type)) {
                window.location.reload();
            }
        } catch (e) {
            console.error('Error parsing message:', e);
        }
    };

    ws.onerror = function(error) {
        console.error('WebSocket Error:', error);
    };

    ws.onclose = function() {
        console.log('WebSocket connection closed');
    };
})();
