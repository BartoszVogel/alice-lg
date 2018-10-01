
import React from 'react'
import {connect} from 'react-redux'
import moment from 'moment'

import RelativeTime from 'components/relativetime'



const RefreshState = function(props) {
  if (!props.cachedAt || !props.cacheTtl) {
    return null;
  }

  const cachedAt = moment.utc(props.cachedAt);
  const cacheTtl = moment.utc(props.cacheTtl);

  if (cacheTtl.isBefore(moment.utc())) {
    // This means cache is currently being rebuilt
    return (
      <li>
        Routes cache was built <b><RelativeTime value={cachedAt} /> </b>
        and is currently being refreshed. 
      </li>
    );

  }

  return (
    <li>
      Routes cache was built <b><RelativeTime value={cachedAt} /> </b>
      and will be refreshed <b><RelativeTime value={cacheTtl} /></b>.
    </li>
  );
}

class ResultsBox extends React.Component {

  render() {
    if (this.props.query == '') {
      return null;
    }

    const queryDuration = this.props.queryDuration.toFixed(2);
    const cachedAt = this.props.cachedAt;
    const cacheTtl = this.props.cacheTtl;

    return (
      <div className="card">
        <div className="lookup-result-summary">
          <ul>
            <li>
              Found <b>{this.props.totalImported}</b> received 
              and <b>{this.props.totalFiltered}</b> filtered routes.
            </li>
            <li>Query took <b>{queryDuration} ms</b> to complete.</li>
            <RefreshState cachedAt={this.props.cachedAt}
                          cacheTtl={this.props.cacheTtl} />
          </ul>
        </div>
      </div>
    );
  }
}


export default connect(
  (state) => {
    return {
      totalImported: state.lookup.totalRoutesImported,
      totalFiltered: state.lookup.totalRoutesFiltered, 

      cachedAt: state.lookup.cachedAt,
      cacheTtl: state.lookup.cacheTtl,

      queryDuration: state.lookup.queryDurationMs
    }
  }
)(ResultsBox)

