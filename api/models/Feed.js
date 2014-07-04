'use strict';

module.exports = {

	schema:true,

  attributes: {
  	name:'string',
  	expires:'integer',	//unix timestamp

  	twitterTextTerms:'array',
  	twitterDomains:'array',
  	twitterHashTags:'array'

  }

};

