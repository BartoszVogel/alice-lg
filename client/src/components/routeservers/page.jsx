
import {debounce} from 'underscore'

import React from 'react'
import {connect} from 'react-redux'

import PageHeader from 'components/page-header'
import Details from './details'
import Status from './status'

import SearchInput from 'components/search-input'

import Protocols from './protocols'

import {setProtocolsFilterValue,
        setProtocolsFilter} from './actions'

class RouteserversPage extends React.Component {

  constructor(props) {
    super(props);
    this.dispatchDebounced = debounce(this.props.dispatch, 350);
  }


  setFilter(value) {
    // Set filter value (for input rendering)
    this.props.dispatch(setProtocolsFilterValue(value));

    // Set filter delayed
    this.dispatchDebounced(setProtocolsFilter(value));

  }

  
  componentDidMount() {
    // Reset Filters
    this.props.dispatch(setProtocolsFilterValue(""));
    this.props.dispatch(setProtocolsFilter(""));
  }


  render() {
    const rsId = this.props.match.params.routeserverId;

    return(
      <div className="routeservers-page">
        <PageHeader>
          <Details routeserverId={rsId} />
        </PageHeader>

        <div className="row details-main">
          <div className="col-lg-9 col-xs-12 col-md-8">
            <div className="card">
              <SearchInput
                value={this.props.protocolsFilterValue}
                placeholder="Filter by Neighbour, ASN or Description"
                onChange={(e) => this.setFilter(e.target.value)}
              />
            </div>

            <Protocols protocol="bgp" routeserverId={rsId} />
          </div>
          <div className="col-lg-3 col-md-4 col-xs-12">
            <div className="card">
              <Status routeserverId={rsId} />
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default connect(
  (state) => {
    return {
      protocolsFilterValue: state.routeservers.protocolsFilterValue
    };
  }
)(RouteserversPage);


