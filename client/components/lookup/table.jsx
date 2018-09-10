
/*
 * Lookup Results Table
 * --------------------
 */

import React from 'react'
import {connect} from 'react-redux'
import {Link} from 'react-router'
import {push} from 'react-router-redux'


import {_lookup,
        ColDefault,
        ColNetwork,
        ColAsPath} from 'components/routeservers/routes/column'

import {showBgpAttributes}
  from 'components/routeservers/routes/bgp-attributes-modal-actions'


// Link Wrappers:
const ColLinkedNeighbor = function(props) {
  const route = props.route;
  const to = `/routeservers/${route.routeserver.id}/protocols/${route.neighbour.id}/routes`;
  
  return (
    <td>
      <Link to={to}>{_lookup(props.route, props.column)}</Link>
    </td>
  );
}

const ColLinkedRouteserver = function(props) {
  const route = props.route;
  const to = `/routeservers/${route.routeserver.id}`;
  
  return (
    <td>
      <Link to={to}>{_lookup(props.route, props.column)}</Link>
    </td>
  );
}


// Custom RouteColumn
const RouteColumn = function(props) {
  const widgets = {
    "network": ColNetwork,

    "bgp.as_path": ColAsPath,
    "ASPath": ColAsPath,

    "neighbour.description": ColLinkedNeighbor,
    "neighbour.asn": ColLinkedNeighbor,
    
    "routeserver.name": ColLinkedRouteserver
  };

  let Widget = widgets[props.column] || ColDefault;
  return (
    <Widget column={props.column} route={props.route}
            displayReasons={props.displayReasons}
            onClick={props.onClick} />
  );
}


class RoutesTable extends React.Component {
  showAttributesModal(route) {
    this.props.dispatch(showBgpAttributes(route));
  }

  render() {
    let routes = this.props.routes;
    const routesColumns = this.props.routesColumns;
    const routesColumnsOrder = this.props.routesColumnsOrder;

    if (!routes || !routes.length) {
      return null;
    }

    let routesView = routes.map((r,i) => {
      return (
        <tr key={`${r.network}_${i}`}>
          {routesColumnsOrder.map(col => {
            return (<RouteColumn key={col}
                                 onClick={() => this.showAttributesModal(r)}
                                 column={col}
                                 route={r}
                                 displayReasons={this.props.type} />);
            }
          )}
        </tr>
      );
    });

    return (
      <table className="table table-striped table-routes">
        <thead>
          <tr>
            {routesColumnsOrder.map(col => <th key={col}>{routesColumns[col]}</th>)}
          </tr>
        </thead>
        <tbody>
          {routesView}
        </tbody>
      </table>
    );
  }
}

export default connect(
  (state) => ({
    routesColumns:      state.config.lookup_columns,
    routesColumnsOrder: state.config.lookup_columns_order,
  })
)(RoutesTable);


