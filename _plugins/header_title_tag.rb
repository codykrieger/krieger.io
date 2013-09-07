module Jekyll
  class HeaderTitleTag < Liquid::Tag
    def initialize(tag_name, text, tokens)
      super
    end

    def render(context)
      title = context.registers[:page]["header_title"]
      return "" if title.nil? || title.empty?
      "<h1>#{title}</h1>"
    end
  end
end

Liquid::Template.register_tag('header_title', Jekyll::HeaderTitleTag)
