'use strict';

/*jshint esnext: true */
/*global angular: false */

import MainCtrl from './mainCtrl';
import LoginCtrl from './loginCtrl';

var app = angular.module('Rr', ['ngRoute'])
          .controller('MainCtrl', MainCtrl)
          .controller('LoginCtrl', LoginCtrl);


app.config(['$routeProvider', $routeProvider => {
	$routeProvider.
		when('/', {
			templateUrl: 'partials/home.html',
			controller: 'MainCtrl as mc'
		}).
		when('/r/:repository', {
			templateUrl: 'partials/repository.html'
		}).
		when('/login', {
			templateUrl: 'partials/login.html',
			controller: 'LoginCtrl as lc'
		}).
		otherwise({
			redirectTo: '/GG.'
		});
}]);