module Jekyll
  class HeaderTitleTag < Liquid::Tag
    def initialize(tag_name, text, tokens)
      super
    end

    def render(context)
      title = context.registers[:page]["header_title"]
      return "" if title.nil? || title.empty?
      "<h2 class='article-or-page-title'>#{title}</h2>"
    end
  end
end

Liquid::Template.register_tag('header_title', Jekyll::HeaderTitleTag)
