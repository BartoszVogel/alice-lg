/*
 * Prefix Lookup Reducer
 */

import {LOAD_RESULTS_REQUEST,
        LOAD_RESULTS_SUCCESS,
        LOAD_RESULTS_ERROR}
 from './actions'

const initialState = {
  query: '',

  results: [],
  error: null,

  queryDurationMs: 0.0,

  limit: 100,
  offset: 0,
  totalRoutes: 0,

  isLoading: false
}

export default function reducer(state=initialState, action) {
  switch(action.type) {
    case LOAD_RESULTS_REQUEST:
      return Object.assign({}, state, initialState, {
        query: action.payload.query,
        isLoading: true
      });
    case LOAD_RESULTS_SUCCESS:
      if (state.query != action.payload.query) {
        return state;
      }

      return Object.assign({}, state, {
        isLoading: false,
        query: action.payload.query,
        queryDurationMs: action.payload.results.query_duration_ms,
        results: action.payload.results.routes,
        limit: action.payload.results.limit,
        offset: action.payload.results.offset,
        totalRoutes: action.payload.results.total_routes,
        error: null
      });

    case LOAD_RESULTS_ERROR:
      if (state.query != action.payload.query) {
        return state;
      }

      return Object.assign({}, state, initialState, {
        query: action.payload.query,
        error: action.payload.error
      });
  }
  return state;
}


