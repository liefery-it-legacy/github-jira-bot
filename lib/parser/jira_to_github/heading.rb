# frozen_string_literal: true

module Parser
  module JiraToGithub
    class Heading
      def call(content)
        content&.gsub(/h(1|2|3|4|5|6)\./) { |h| "#" * h[1].to_i }
      end
    end
  end
end
