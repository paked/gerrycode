import { HomeView, RepositoriesView } from './views';

class Router extends Backbone.Router {

  constructor () {
    this.routes = {
      '': 'home',
      'repositories': 'repositories'
    };
    super();
  }

  home () {
    console.log('Route#home');
    var view = new HomeView();
    $('#app').html(view.render().$el);
  }

  repositories () {
    console.log('Route#resources');
    var view = new RepositoriesView();
    $('#app').html(view.render().$el);
  }

}

export default Router;