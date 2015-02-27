app = angular.module('revco', ["ngRoute"]);

app.config(['$routeProvider',
	function($routeProvider) {
		$routeProvider.
			when('/', {
				templateUrl: '/partials/home.html',
				controller: 'MainCtrl'
			}).
			when('/view/:project', {
				templateUrl: '/partials/view_project.html'
			}).
			when('/explore', {
				templateUrl: '/partials/explore.html'
			}).
			otherwise({
				redirectTo: '/'
			});

	}]);

app.controller('MainCtrl', function($scope) {
	$scope.message = "PHONETAB EAT PHONETAB";
});
