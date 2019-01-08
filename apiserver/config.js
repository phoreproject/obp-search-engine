module.exports = {
  'name': 'Phore Search',

  'logo': 'https://search.phore.io/logo.png',

  'links': {
    'self': '',
    'listings': 'https://search.phore.io/search/listings'
  },

  'options': {
    'rating': {
      'type': 'radio',
      'label': 'Rating',
      'options': [
        {
          'value': '5',
          'label': '⭐⭐⭐⭐⭐',
          'checked': true,
          'default': true
        },
        {
          'value': '4',
          'label': '⭐⭐⭐⭐ \u0026 up',
          'checked': false,
          'default': false
        },
        {
          'value': '3',
          'label': '⭐⭐⭐ \u0026 up',
          'checked': false,
          'default': false
        },
        {
          'value': '2',
          'label': '⭐⭐ \u0026 up',
          'checked': false,
          'default': false
        },
        {
          'value': '1',
          'label': '⭐ \u0026 up',
          'checked': false,
          'default': false
        }
      ]
    },
  },

  'sortBy': {
    'PRICE_ASC': {
      'label': 'Price (Low to High)',
      'selected': false,
      'default': false
    },
    'PRICE_DESC': {
      'label': 'Price (High to Low)',
      'selected': false,
      'default': false
    },
    'NAME_ASC': {
      'label': 'Name (Ascending)',
      'selected': false,
      'default': false
    },
    'NAME_DESC': {
      'label': 'Name (Descending)',
      'selected': false,
      'default': false
    },
    'RATING_ASC': {
      'label': 'Rating (Low to High)',
      'selected': false,
      'default': false
    },
    'RATING_DESC': {
      'label': 'Rating (High to Low)',
      'selected': false,
      'default': false
    },
    'RELEVANCE': {
      'label': 'Relevance',
      'select': true,
      'default': true
    }
  }
};
