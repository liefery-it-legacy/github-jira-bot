# frozen_string_literal: true

module Parser
  class Image
    def call(content, format:)
      github(content) if format == 'github'
    end

    private

    def github(content)
      content&.gsub(/(?:!\[(.*?)\]\((.*?)\))/) { |image| "!#{image.split('(').last.delete(')')}!" }
    end
  end
end
