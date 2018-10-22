require "http/server"
require "blank"

server = HTTP::Server.new("0.0.0.0", 8080, [HTTP::LogHandler.new]) do |context|
  context.response.content_type = "text/plain"
  context.response.print "IsBlank: #{"".blank?}"
end

puts "Listening on port 8080"
server.listen
