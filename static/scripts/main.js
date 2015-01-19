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