# frozen_string_literal: true

module Parser
  class Heading
    def call(content, format:)
      if format == 'github'
        github(content)
      elsif format == 'jira'
        jira(content)
      end
    end

    private

    def github(content)
      content&.gsub(/(\#\#\#\#\#\#|\#\#\#\#\#|\#\#\#\#|\#\#\#|\#\#|\#)/) { |h| "h#{h.length}." }
    end

    def jira(content)
      content&.gsub(/h(1|2|3|4|5|6)\./) { |h| "#" * h[1].to_i }
    end
  end
end
