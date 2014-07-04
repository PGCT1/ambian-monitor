'use strict';

var config = require('../../config.json');

module.exports = function(req, res, next) {
	
  if (req.header('apiKey') == config.apiKey) {
    return next();
  }
  
  return res.unauthorized('You are not permitted to perform this action.');
};
