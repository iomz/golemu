var waitAndSend = function(message, callback) {
    waitForConnection(function() {
        ws.send(message);
        if (typeof callback !== "undefined") {
            callback();
        }
    }, 1e3);
};

var waitForConnection = function(callback, interval) {
    if (ws.readyState === 1) {
        callback();
    } else {
        var that = this;
        setTimeout(function() {
            that.waitForConnection(callback, interval);
        }, interval);
    }
};

var retrieveTagList = function() {
    var retrieve_tag = {
        updateType: "retrieve",
        tag: {
            pcBits: "",
            length: "",
            epcLengthBits: "",
            epc: "",
            readData: ""
        }
    };
    waitAndSend(JSON.stringify(retrieve_tag));
};

$("#textin").val("");

$("#send").click(function(event) {
    ws.send($("#textin").val());
    $("#textin").val("");
});

$("#add-tag").click(function(event) {
    var add_tag = {
        updateType: "add",
        tag: {
            pcBits: "29a9",
            length: "16",
            epcLengthBits: "80",
            epc: "dc20420c4c72cf4d76de",
            readData: "1f1f"
        }
    };
    ws.send(JSON.stringify(add_tag));
});

$("#delete-tag").click(function(event) {
    var delete_tag = {
        updateType: "delete",
        tag: {
            pcBits: "29a9",
            length: "16",
            epcLengthBits: "80",
            epc: "dc20420c4c72cf4d76de",
            readData: "1f1f"
        }
    };
    ws.send(JSON.stringify(delete_tag));
});

$("#retrieve-tag").click(function(event) {
    retrieveTagList();
});

try {
    var ws = new WebSocket("ws://" + window.location.host + "/ws");
    console.log("Websocket - status: " + ws.readyState);
    ws.onopen = function(m) {
        console.log("CONNECTION opened..." + this.readyState);
    };
    retrieveTagList();
    ws.onmessage = function(m) {
        console.log(JSON.parse(m.data));
    };
    ws.onerror = function(m) {
        console.log("Error occured sending..." + m.data);
    };
    ws.onclose = function(m) {
        console.log("Disconnected - status " + this.readyState);
    };
} catch (exception) {
    console.log(exception);
}