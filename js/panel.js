var myApp = angular.module('myApp',[]);

$(document).ready(function(){
	// Set article height
 	var winH = $(window).height();
 	var footerH = $('#footer').height();
 	$('.content article').css('min-height',winH-footerH+'px');
});
