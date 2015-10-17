"use strict;"

$(document).ready ( function(){
var margin = {top: 20, right: 30, bottom: 30, left: 40},
    width = 1160 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

    function render(data) {
      var x = d3.scale.linear()
                      .domain(d3.extent(data, function(d) { return d.id; }))
                      .range([0, width]);

      var y = d3.scale.linear()
                      .domain([
                        d3.min(data,function(d){return d.days;}),
                        d3.max(data,function(d){return d.days;}),
                      ])
                      .range([height,0]);

      var xAxis = d3.svg.axis()
          .scale(x)
          .orient("bottom");

      var yAxis = d3.svg.axis()
          .scale(y)
          .orient("left");

      chart.append("g")
          .attr("class", "x axis")
          .attr("transform", "translate(0," + height + ")")
          .call(xAxis);

      chart.append("g")
          .attr("class", "y axis")
          .call(yAxis);

      chart.selectAll(".bar")
          .data(data)
        .enter().append("rect")
          .attr("class", "bar")
          .attr("x", function(d) { return x(d.id); })
          .attr("y", function(d) { return y(Math.max(0, d.days)); })
          .attr("height", function(d) { return Math.abs(y(d.days) - y(0)); })
          .attr("width", 2 )
          .on("click", function(d,i) {
            window.open("https://wunderlist.com/#/tasks/" + d.id);
          })
          .on("mouseover", function(d) {
            d3.select(this)
            .transition().duration(200)
            .attr("width", 10)
            .style("fill-opacity", 0.5)
          })
          .on("mouseout", function(d,i) {
            d3.select(this)
            .transition().duration(200)
            .attr("width", 2)
            .style("fill-opacity", 1)
          })
          .append("svg:title")
          .text(function(d) { return d.id + "," + d.days; })
          ;
  };

    var chart = d3.select(".chart")
        .attr("width", width + margin.left + margin.right)
        .attr("height", height + margin.top + margin.bottom)
        .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    d3.json("/api/v1/tasks", render);
});
