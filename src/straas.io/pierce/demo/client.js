// pls npm i socket.io-client
var socketio = require('socket.io-client');

var socketUrl = "http://localhost:5000/socket.io";
var log = console.log;

// connect to bonjour to receive task script
function runSocket() {

  var socket = socketio.connect(socketUrl, {
    transports: ['websocket'],
    'reconnection': false,
    'forceNew': true,
    'reconnectionDelayMax': 3000,
    'query': 'role=agent',
  });

  socket.on('connect_error', function(err) {
    log('connect_error', err);
  });

  socket.on('connect', function() {
    log('connect');
  });

  socket.on('error', function(err) {
    logr('error', err);
  });

  socket.on('connect_error', function(err) {
    log('connect_error', err);
  });

  socket.on('connect_timeout', function(err) {
    log('connect_timeout', err);
  });

  socket.on('reconnect', function() {
    log('reconnect');
  });

  socket.on('data', function() {
    log(arguments);
  });

}

runSocket();
process.stdin.resume();
