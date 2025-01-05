document.addEventListener('DOMContentLoaded', function() {
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

    // Flag to track if a link click is in progress
    let clickInProgress = false;

    // Capture clicks on <a> tags to prevent reloading appening when clicking on link
    document.addEventListener('click', function(event) {
        const target = event.target;
        if (target.tagName === 'A' && target.href) {
            clickInProgress = true;
            // Wait a bit before allowing reload again
            setTimeout(function() {
                clickInProgress = false;
            }, 5000);
        }
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

            // Reload the page immediately if we're not in the grace period and no clicks are in progress
            if (!gracePeriod && !clickInProgress) {
                window.location.reload();
                return;
            }

            // If still in the grace period or a click is in progress, debounce the reload
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(function() {
                if (!clickInProgress) {
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
