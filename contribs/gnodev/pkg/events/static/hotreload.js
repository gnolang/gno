(function() {
    // Define the events that will trigger a page reload
    var eventsReload = {{ .ReloadEvents | jsEventsArray }};
    
    // Establish the WebSocket connection to the event server
    var ws = new WebSocket('ws://{{- .Remote -}}');
    
    // Flag to determine if the page is in the grace period after loading
    var gracePeriod = true;
    var graceTimeout = 1000; // ms

    // Set a timer to end the grace period after `graceTimeout`
    var debounceTimeout = setTimeout(function() {
        gracePeriod = false;
    }, graceTimeout); 

    // Handle incoming WebSocket messages
    ws.onmessage = function(event) {
        try {
            var message = JSON.parse(event.data);
            console.log('Receiving message:', message);

            // Ignore events not in the reload-triggering list
            if (!eventsReload.includes(message.type)) {
                return; 
            }

            // Reload the page immediately if we're not in the grace period
            if (!gracePeriod) {
                window.location.reload();
                return;
            }

            // If still in the grace period, debounce the reload
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(function() {
                window.location.reload();
            }, graceTimeout);

        } catch (e) {
            console.error('Error handling message:', e);
        }
    };

    // Handle ws errors and closure
    ws.onerror = function(error) {
        console.error('WebSocket Error:', error);
    };
    ws.onclose = function() {
        console.log('WebSocket connection closed');
    };
})();
