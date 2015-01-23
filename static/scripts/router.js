import { HomeView, RepositoriesView, LoginView } from './views';

class Router extends Backbone.Router {

  constructor () {
    console.log("PRINTING")
    console.log("__PRINTING__")
    this.routes = {
      '': 'home',
      'repositories': 'repositories',
      'login':'login'
    };
    super();
  }

  home () {
    if (window.token == "" || window.token == undefined){
      $(location).attr("href", "#/login")
      console.log("Yeah go to login m8")
      return
    }

    console.log(window.token)
    console.log('Route#home');
    var view = new HomeView();
    $('#app').html(view.render().$el);
  }

  repositories () {
    console.log('Route#resources');
    var view = new RepositoriesView();
    $('#app').html(view.render().$el);
  }

  login() {
  	console.log("Route#login")
  	var view = new LoginView()
  	$('#app').html(view.render().$el)
  }

}

export default Router;