#!/bin/env ruby

origin = [176/2, 176/2]
r = 500 * 176 / 1080
points = []
n = 60

(1..n).each do |i|
  x = origin[0] + (r * Math.sin((i*2*Math::PI + 1.5) / n))
  y = origin[1] + (r * Math.cos((i*2*Math::PI) / n))
  points << [x.floor, y.floor]
end

puts "circles := [#{n}][2]int16 {"
points.reverse.each do |p|
  puts "{#{p[0]}, #{p[1]}},"
end
puts "}"