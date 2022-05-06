class Renderer {
    rawData;

    constructor(rawData) {
        this.rawData = rawData;
    }

    renderUsernames() {
        this.rawData = this.rawData.replace(/( |\n)@([_a-z0-9]{5,16})/, "$1[@$2](/r/users:$2)")
    }

    renderAll() {
        this.renderUsernames();
        return this.rawData;
    }
}

