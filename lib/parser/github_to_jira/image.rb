# frozen_string_literal: true

module Parser
  module GithubToJira
    class Image
      def call(content)
        content&.gsub(/(?:!\[(.*?)\]\((.*?)\))/) { |image| "!#{image.split('(').last.delete(')')}!" }
      end
    end
  end
end
