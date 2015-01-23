import Router from './router';

class Application {
	constructor () {
		new Router();
		Backbone.history.start();
	}
}

$(() => {
  new Application();
});

window.token = ""

console.log("loading rr")
// console.log("loaded rr")