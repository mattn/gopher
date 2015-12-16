$(function() {
  var ws = new WebSocket("ws://" + location.host + "/chat");
  ws.onerror = function(m) {
    console.log("Error occured: " + m.data);
  };
  ws.onmessage = function(m) {
    var msg = JSON.parse(m.data);
    switch (msg.type) {
    case 'whoami': $('#whoami').text(msg.user); break;
    case 'message': $('#chatbox').append($('<p/>').text(msg.user + ": " + msg.value)); break;
    }
  };
  $('#send').click(function() {
    ws.send(JSON.stringify({type: 'message', value: $('#text').val()}));
    $('#text').val("");
  });
});
// vim:set et ts=2:
