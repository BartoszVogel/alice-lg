
import React from 'react'
import {connect} from 'react-redux'

import PageHeader from 'components/page-header'

import Lookup from 'components/lookup'
import LookupSummary from 'components/lookup/results-summary'

import Content from 'components/content'

class _LookupView extends React.Component {
  render() {
    if (this.props.enabled == false) {
      return null;
    }

    return (
      <div className="lookup-container">
       <div className="col-md-8">
         <Lookup />
       </div>
       <div className="col-md-4">
         <LookupSummary />
       </div>
      </div>
    );
  }
}

const LookupView = connect(
  (state) => {
    return {
      enabled: state.config.prefix_lookup_enabled
    }
  }
)(_LookupView);


export default class LookupPage extends React.Component {
  render() {
    return (
      <div className="welcome-page">
       <PageHeader></PageHeader>
       <p></p>
       <LookupView />
      </div>
    );
  }
}

