
import React from 'react'
import {connect} from 'react-redux'

import {Link} from 'react-router'


const PageLink = function(props) {
/*
  const linkPage = parseInt(props.page);

  let pr = props.pageReceived;
  let pf = props.pageFiltered;
  let pn = props.pageNotExported;

  // This here can be surely more elegant.
  switch(props.anchor) {
    case "routes-received":
      pr = linkPage;
      break;
    case "routes-filtered":
      pf = linkPage;
      break;
    case "routes-not-exported":
      pn = linkPage;
      break;
  }

  const search = `?pr=${pr}&pf=${pf}&pn=${pn}`;
  const hash   = `#${props.anchor}`;
  const linkTo = {
    pathname: routing.pathname,
    hash:     hash,
  };

  return (
    <Link to={linkTo}>{linkPage + 1}</Link>
  );
  */
  return (<p>23</p>);
};


class RoutesPaginatorView extends React.Component {

  render() {
    if (this.props.totalPages <= 1) {
      return null; // Nothing to paginate
    }

    const pageLinks = Array.from(Array(this.props.totalPages), (_, i) => {
        /*
               <pagelink page={i}
                              anchor={this.props.anchor} />
          */
      return (
        <li key={i}>
            {i + 1}
        </li>
      );
    });


    return (
      <nav aria-label="Routes Pagination">
        <ul className="pagination">
          <li>
              <a href="#" aria-label="Previous">
                <span aria-hidden="true">&laquo;</span>
              </a>
          </li>
          <li>
            <PageLink page={0} anchor={this.props.anchor} />
          </li>

          {pageLinks}
          <li>
              <a href="#" aria-label="Next">
                <span aria-hidden="true">&raquo;</span>
              </a>
           </li>
        </ul>
      </nav>
    );
  }
}

export const RoutesPagniator = connect(
  (state) => ({
      pageReceived:    state.routes.receivedPage,
      pageFiltered:    state.routes.filteredPage,
      pageNotExported: state.routes.notExportedPage,

      routing: state.routing.locationBeforeTransitions
  })
)(RoutesPaginatorView);


export class RoutesPaginationInfo extends React.Component {
  render() {
    const totalResults = this.props.totalResults;
    const perPage = this.props.pageSize;
    const start = this.props.page * perPage + 1;
    const end = Math.min(start + perPage - 1, totalResults);
    if (this.props.totalPages == 1) {
      let routes = "route";
      if (totalResults > 1) {
        routes = "routes";
      }

      return (
        <div className="routes-pagination-info pull-right">
          Showing <b>all</b> of <b>{totalResults}</b> {routes}
        </div>
      );
    }
    return (
      <div className="routes-pagination-info pull-right">
        Showing <b>{start} - {end}</b> of <b>{totalResults}</b> total routes
      </div>
     );
  }
}

