var tagTile, isWaiting = false;

String.prototype.hashCode = function() {
    var hash = 0;
    if (this.length == 0) return hash;
    for (i = 0; i < this.length; i++) {
        char = this.charCodeAt(i);
        hash = (hash << 5) - hash + char;
        hash = hash & hash;
    }
    return hash;
};

var waitAndSend = function(message, callback) {
    waitForConnection(function() {
        ws.send(message);
        if (typeof callback !== "undefined") {
            callback();
        }
    }, 100);
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
    isWaiting = true;
};

var addTag = function(t) {
    var newBr1 = $("<br/>", {});
    var newBr2 = $("<br/>", {});
    var newH2 = $("<h2/>", {
        text: t.Length + "/" + t.EPCLengthBits + "/" + t.PCBits + "/" + t.ReadData
    });
    var newImageOverlay = $("<div/>", {
        "class": "image-overlay"
    });
    newBr1.appendTo(newImageOverlay);
    newBr2.appendTo(newImageOverlay);
    newH2.appendTo(newImageOverlay);
    var newImageContainer = $("<div/>", {
        "class": "image-container"
    });
    newImageOverlay.appendTo(newImageContainer);
    var newTileContent = $("<div/>", {
        "class": "tile-content"
    });
    newImageContainer.appendTo(newTileContent);
    var newTileLabel = $("<span/>", {
        "class": "tile-label",
        text: t.EPC
    });
    var tagString = t.EPC + t.Length + t.EPCLengthBits + t.PCBits + t.ReadData;
    var bgColor = "bg-darkBlue";
    switch (parseInt(t.EPCLengthBits)) {
      case 80:
        bgColor = "bg-teal";
        break;

      case 96:
        bgColor = "bg-magenta";
        break;

      case 128:
        bgColor = "bg-orange";
        break;
    }
    var newTile = $("<div/>", {
        "class": "tile-wide fg-white tag-tile " + bgColor,
        "data-role": "tile",
        id: tagString.hashCode()
    });
    newTileContent.appendTo(newTile);
    newTileLabel.appendTo(newTile);
    $("#tag-cloud").append(newTile);
    setTimeout(function() {
        newTile.css({
            opacity: 1,
            "-webkit-transform": "scale(1)",
            transform: "scale(1)",
            "-webkit-transition": ".3s",
            transition: ".3s"
        });
    }, Math.floor(Math.random() * 500));
};

var addTagFromDialog = function() {
    var tagToAdd = {
        UpdateType: "add",
        Tag: {
            PCBits: $("#PCBits").val(),
            Length: $("#Length").val(),
            EPCLengthBits: $("#EPCLengthBits").val(),
            EPC: $("#EPC").val(),
            ReadData: $("#ReadData").val()
        },
        Tags: null
    };
    ws.send(JSON.stringify(tagToAdd));
    hideMetroDialog("#dialog");
    isWaiting = true;
};

var deleteTag = function(t) {
    var tagString = t.EPC + t.Length + t.EPCLengthBits + t.PCBits + t.ReadData;
    $("#" + tagString.hashCode()).remove();
};

var deleteTagFromDialog = function() {
    var tagToDelete = {
        UpdateType: "delete",
        Tag: {
            PCBits: $("#PCBits").val(),
            Length: $("#Length").val(),
            EPCLengthBits: $("#EPCLengthBits").val(),
            EPC: $("#EPC").val(),
            ReadData: $("#ReadData").val()
        },
        Tags: null
    };
    ws.send(JSON.stringify(tagToDelete));
    hideMetroDialog("#dialog");
    isWaiting = true;
};

var updateTagFromDialog = function(t) {
    hideMetroDialog("#dialog");
    notifyOnError();
};

var editTag = function(t) {
    $("#EPC").val(t.EPC);
    $("#PCBits").val(t.PCBits);
    $("#Length").val(t.Length);
    $("#EPCLengthBits").val(t.EPCLengthBits);
    $("#ReadData").val(t.ReadData);
    $("#rand-epc-btn").show();
    $("#rand-iso-btn").show();
    $("#rand-prop-btn").show();
    $("#add-btn").hide();
    $("#update-btn").show();
    $("#delete-btn").show();
    showDialog("#dialog");
};

var showDialog = function(id) {
    var dialog = $(id).data("dialog");
    dialog.open();
};

function notifyOnSuccess(m) {
    var action = "";
    switch (m.UpdateType) {
      case "add":
        action = "Added";
        break;

      case "delete":
        action = "Deleted";
        break;

      default:
        return;
    }
    $.Notify({
        caption: "Success",
        content: action + " a tag: " + m.Tag.EPC,
        type: "success"
    });
}

function notifyOnError() {
    $.Notify({
        caption: "Error",
        content: "Something went wrong",
        type: "alert"
    });
}

$("#add-tile").click(function(event) {
    $("#EPC").val("");
    $("#PCBits").val("");
    $("#Length").val("");
    $("#EPCLengthBits").val("");
    $("#ReadData").val("");
    $("#rand-epc-btn").show();
    $("#rand-iso-btn").show();
    $("#rand-prop-btn").show();
    $("#add-btn").show();
    $("#update-btn").hide();
    $("#delete-btn").hide();
    showDialog("#dialog");
});

$("#delete-tile").click(function(event) {
    $("#EPC").val("");
    $("#PCBits").val("");
    $("#Length").val("");
    $("#EPCLengthBits").val("");
    $("#ReadData").val("");
    $("#rand-epc-btn").hide();
    $("#rand-iso-btn").hide();
    $("#rand-prop-btn").hide();
    $("#add-btn").hide();
    $("#update-btn").hide();
    $("#delete-btn").show();
    showDialog("#dialog");
});

$("#retrieve-tag").click(function(event) {
    retrieveTagList();
});

$(document).click(function(e) {
    var src = $(e.target);
    tagTile = src.parents(".tag-tile");
    if (tagTile.length != 0) {
        var epc = tagTile.children(".tile-label").text();
        var info = $("h2", tagTile).text().split("/");
        editTag({
            PCBits: info[2],
            Length: info[0],
            EPCLengthBits: info[1],
            EPC: epc,
            ReadData: info[3]
        });
    }
});

try {
    var ws = new WebSocket("ws://" + window.location.host + "/ws");
    console.log("Websocket - status: " + ws.readyState);
    ws.onopen = function(m) {
        console.log("CONNECTION opened..." + this.readyState);
        retrieveTagList();
    };
    ws.onmessage = function(m) {
        var m = JSON.parse(m.data);
        switch (m.UpdateType) {
          case "add":
            addTag(m.Tag);
            if (isWaiting) {
                notifyOnSuccess(m);
                isWaiting = false;
            }
            break;

          case "delete":
            deleteTag(m.Tag);
            if (isWaiting) {
                notifyOnSuccess(m);
                isWaiting = false;
            }
            break;

          case "retrieval":
            for (var i = 0; i < m.Tags.length; i++) {
                addTag(m.Tags[i]);
            }
            break;

          case "error":
            notifyOnError();
            break;

          default:        }
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
