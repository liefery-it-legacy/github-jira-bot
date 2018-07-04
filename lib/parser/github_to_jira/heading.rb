# frozen_string_literal: true

module Parser
  module GithubToJira
    class Heading
      def call(content)
        content&.gsub(/(\#\#\#\#\#\#|\#\#\#\#\#|\#\#\#\#|\#\#\#|\#\#|\#)/) { |h| "h#{h.length}." }
      end
    end
  end
end
