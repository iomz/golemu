function showCharms(id) {
    var charm = $(id).data("charm");
    if (charm.element.data("opened") === true) {
        charm.close();
    } else {
        charm.open();
    }
}

$(function() {
    var tiles = $(".tile, .tile-small, .tile-sqaure, .tile-wide, .tile-large, .tile-big, .tile-super");
    $.each(tiles, function() {
        var tile = $(this);
        setTimeout(function() {
            tile.css({
                opacity: 1,
                "-webkit-transform": "scale(1)",
                transform: "scale(1)",
                "-webkit-transition": ".3s",
                transition: ".3s"
            });
        }, Math.floor(Math.random() * 500));
    });
    $(".tile-group").animate({
        left: 0
    });
});

$(function() {
    var current_tile_area_scheme = localStorage.getItem("tile-area-scheme") || "tile-area-scheme-dark";
    $(".tile-area").removeClass(function(index, css) {
        return (css.match(/(^|\s)tile-area-scheme-\S+/g) || []).join(" ");
    }).addClass(current_tile_area_scheme);
    $(".schemeButtons .button").hover(function() {
        var b = $(this);
        var scheme = "tile-area-scheme-" + b.data("scheme");
        $(".tile-area").removeClass(function(index, css) {
            return (css.match(/(^|\s)tile-area-scheme-\S+/g) || []).join(" ");
        }).addClass(scheme);
    }, function() {
        $(".tile-area").removeClass(function(index, css) {
            return (css.match(/(^|\s)tile-area-scheme-\S+/g) || []).join(" ");
        }).addClass(current_tile_area_scheme);
    });
    $(".schemeButtons .button").on("click", function() {
        var b = $(this);
        var scheme = "tile-area-scheme-" + b.data("scheme");
        $(".tile-area").removeClass(function(index, css) {
            return (css.match(/(^|\s)tile-area-scheme-\S+/g) || []).join(" ");
        }).addClass(scheme);
        current_tile_area_scheme = scheme;
        localStorage.setItem("tile-area-scheme", scheme);
        showSettings();
    });
});