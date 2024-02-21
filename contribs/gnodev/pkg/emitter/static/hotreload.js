(function() {
    // Define the events that will trigger a page reload
    const eventsReload = {{ .ReloadEvents | json }};
    
    // Establish the WebSocket connection to the event server
    const ws = new WebSocket('ws://{{- .Remote -}}');
    
    // `gracePeriod` mitigates reload loops due to excessive events. This period
    // occurs post-loading and lasts for the `graceTimeout` duration.
    const graceTimeout = 1000; // ms
    let gracePeriod = true;
    let debounceTimeout = setTimeout(function() {
        gracePeriod = false;
    }, graceTimeout); 

    // Handle incoming WebSocket messages
    ws.onmessage = function(event) {
        try {
            const message = JSON.parse(event.data);
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
