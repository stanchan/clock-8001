#!/usr/bin/env ruby

origin = [192/2, 192/2]
r = 500 * 192 / 1080
points = []
n = 60

(1..n).each do |i|
  x = origin[0] - (r * Math.sin((-i*2*Math::PI) / n))
  y = origin[1] - (r * Math.cos((-i*2*Math::PI) / n))
  points << [x.round, y.round]
end

puts "circles := [#{n}][2]int16 {"
points.each do |p|
	puts "{#{p[0]}, #{p[1]}},"
end
puts "}"