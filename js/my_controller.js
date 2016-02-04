myApp.controller('myCtrl', function ($scope) {
	
	$scope.hash =  window.location.hash.substr(1);
	if(!$scope.hash){$scope.hash = 'mulifs--music-library-filesystem'}


	$scope.$watch('langCode', function(newValue, oldValue){
		updateLang(newValue);
	});
	
});