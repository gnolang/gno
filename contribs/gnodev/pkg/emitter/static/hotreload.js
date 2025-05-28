document.addEventListener('DOMContentLoaded', function() {
    // Define the events that will trigger a page reload
    const eventsReload = [
        {{range .ReloadEvents}}'{{.}}',{{end}}
    ];

    // Establish the WebSocket connection to the event server
    const ws = new WebSocket('ws://{{- .Remote -}}');

    // `gracePeriod` mitigates reload loops due to excessive events. This period
    // occurs post-loading and lasts for the `graceTimeout` duration.
    const graceTimeout = 1000; // ms
    let gracePeriod = true;
    let debounceTimeout = setTimeout(function() {
        gracePeriod = false;
    }, graceTimeout);

    // Flag to track if navigation is in progress, make sure navigation is not blocked
    let isNavigatingAway = false;

    // This function is called before the page is redirected to another URL
    window.addEventListener('beforeunload', function(event) {
        isNavigatingAway = true;
    });

    // Handle incoming WebSocket messages
    ws.onmessage = function(event) {
        try {
            const message = JSON.parse(event.data);
            console.log('Receiving message:', message);

            // Ignore events not in the reload-triggering list
            if (!eventsReload.includes(message.type)) {
                return;
            }

            // Reload the page immediately if we're not in the grace period and no navigation is in progress.
            if (!gracePeriod && !isNavigatingAway) {
                window.location.reload();
                return;
            }

            // If still in the grace period or a click is in progress, debounce the reload
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(function() {
                if (!isNavigatingAway) {
                    window.location.reload();
                }
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
});
