'use strict';

/*jshint esnext: true */
/*global angular: false */

import MainCtrl from './mainCtrl';

var app = angular.module('Rr', ['ngRoute'])
          .controller('MainCtrl', MainCtrl);


app.config(['$routeProvider', $routeProvider => {
	$routeProvider.
		when('/', {
			templateUrl: 'partials/home.html'
		}).
		when('/r/:repository', {
			templateUrl: 'partials/repository.html'
		}).
		otherwise({
			redirectTo: '/GG.'
		});
}]);