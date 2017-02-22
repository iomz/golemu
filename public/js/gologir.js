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

// converts binary string to a hexadecimal string
// returns an object with key 'valid' to a boolean value, indicating
// if the string is a valid binary string.
// If 'valid' is true, the converted hex string can be obtained by
// the 'result' key of the returned object
function binaryToHex(s) {
    var i, k, part, accum, ret = '';
    for (i = s.length-1; i >= 3; i -= 4) {
        // extract out in substrings of 4 and convert to hex
        part = s.substr(i+1-4, 4);
        accum = 0;
        for (k = 0; k < 4; k += 1) {
            if (part[k] !== '0' && part[k] !== '1') {
                // invalid character
                return { valid: false };
            }
            // compute the length 4 substring
            accum = accum * 2 + parseInt(part[k], 10);
        }
        if (accum >= 10) {
            // 'A' to 'F'
            ret = String.fromCharCode(accum - 10 + 'A'.charCodeAt(0)) + ret;
        } else {
            // '0' to '9'
            ret = String(accum) + ret;
        }
    }
    // remaining characters, i = 0, 1, or 2
    if (i >= 0) {
        accum = 0;
        // convert from front
        for (k = 0; k <= i; k += 1) {
            if (s[k] !== '0' && s[k] !== '1') {
                return { valid: false };
            }
            accum = accum * 2 + parseInt(s[k], 10);
        }
        // 3 bits, value cannot exceed 2^3 - 1 = 7, just convert
        ret = String(accum) + ret;
    }
    return { valid: true, result: ret };
}

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

var updateTagFromDialog = function() {
    var targetTagId = $("#tag-selected").val();
    console.log(targetTagId);
    var targetTagTile = $("#"+targetTagId);
    var epc = targetTagTile.children(".tile-label").text();
    var info = $("h2", targetTagTile).text().split("/");

    var tagToDelete = {
        UpdateType: "delete",
        Tag: {
            PCBits: info[2],
            Length: info[0],
            EPCLengthBits: info[1],
            EPC: epc,
            ReadData: info[3]
        },
        Tags: null
    };
    ws.send(JSON.stringify(tagToDelete));

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
    $("#dialog > form > h1").text("Edit Tag");
    $("#tag-selected").val(t.id);
    showDialog("#dialog");
};

var showDialog = function(id) {
    var dialog = $(id).data("dialog");
    dialog.open();
};

var notifyOnSuccess = function(m) {
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

var notifyOnError = function() {
    $.Notify({
        caption: "Error",
        content: "Something went wrong",
        type: "alert"
    });
};

var makeRandomHexEPCPrefix = function() {
    // First 14bits
    var header = "00110000001101";
    var prefix_list = [
        "11011001111110010001110010011", // Lab01
        "11011010110110010111001010100", // Lab02
        "11011001100011001101000000000", // Daiwa
        "10010110100010101010001000000"  // Masuizumi
    ];
    return binaryToHex(header+prefix_list[Math.floor(Math.random() * prefix_list.length)]);
};

var makeRandomHexString = function(n) {
    // Ensure the resulting string has at least the length of 1
    if (n != parseInt(n, 10) || n <= 0) {
        n = 1;
    }

    var str = "";
    var c = "1234567890abcdef";

    for (var i = 0; i < n; i++) {
        str += c.charAt(Math.floor(Math.random() * c.length));
    }

    return str;
}

var randomFillDialog = function(code) {
    var epc, pcBits, length, epcLengthBits, readData;
    switch (code) {
        case "epc":
            //epc = "302db319a000" + makeRandomHexString(12);
            epc = makeRandomHexEPCPrefix() + makeRandomHexString(13);
            pcBits = "3000";
            length = "18";
            epcLengthBits = "96";
            readData = makeRandomHexString(4);
            break;

        case "iso":
            var isos = [{epcPrefix: "dc20420c4c36", pcBits: "29a9", length: "16", epcLengthBits: "80"},{epcPrefix: "c4a301c70d36cb32920b1dc1", pcBits: "41a2", length: "22", epcLengthBits: "128"}];
            var choice = Math.floor(Math.random() * isos.length);

            epc = isos[choice]["epcPrefix"] + makeRandomHexString(8);
            pcBits = isos[choice]["pcBits"];
            length = isos[choice]["length"];
            epcLengthBits = isos[choice]["epcLengthBits"];
            readData = makeRandomHexString(4);
            break;

        case "proprietary":
            var props = [{words: 32, pcBits: "4000", length: "22", epcLengthBits: "128"},{words: 24, pcBits: "3000", length: "18", epcLengthBits: "96"}];
            var choice = Math.floor(Math.random() * props.length);

            epc = makeRandomHexString(props[choice]["words"]);
            pcBits = props[choice]["pcBits"];
            length = props[choice]["length"];
            epcLengthBits = props[choice]["epcLengthBits"];
            readData = makeRandomHexString(4);
            break;
    }
    $("#EPC").val(epc);
    $("#PCBits").val(pcBits);
    $("#Length").val(length);
    $("#EPCLengthBits").val(epcLengthBits);
    $("#ReadData").val(readData);
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
    $("#dialog > form > h1").text("Add Tag");
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
    $("#dialog > form > h1").text("Delete Tag");
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
            ReadData: info[3],
            id: tagTile.attr("id")
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
        console.log(m);
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
