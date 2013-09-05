module Jekyll
  class TitleTag < Liquid::Tag
    def initialize(tag_name, text, tokens)
      super
    end

    def render(context)
      site_title = context.registers[:site].config["name"]
      page_title = context.registers[:page]["title"]

      return site_title if page_title.nil? || page_title.empty?
      "#{page_title} &mdash; #{site_title}"
    end
  end
end

Liquid::Template.register_tag('title', Jekyll::TitleTag)
