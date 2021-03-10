# frozen_string_literal: true

require 'rails_helper'

RSpec.describe AddressHelper, type: :helper do
  describe '#states_for_select' do
    it 'formats states into an array of full names' do
      expect(helper.states_for_select).to eq(
        [
          ['Alaska', 'AK'],
          ['Alabama', 'AL'],
          ['Arkansas', 'AR'],
          ['Arizona', 'AZ'],
          ['California', 'CA'],
          ['Colorado', 'CO'],
          ['Connecticut', 'CT'],
          ['District of Columbia', 'DC'],
          ['Delaware', 'DE'],
          ['Florida', 'FL'],
          ['Georgia', 'GA'],
          ['Hawaii', 'HI'],
          ['Iowa', 'IA'],
          ['Idaho', 'ID'],
          ['Illinois', 'IL'],
          ['Indiana', 'IN'],
          ['Kansas', 'KS'],
          ['Kentucky', 'KY'],
          ['Louisiana', 'LA'],
          ['Massachusetts', 'MA'],
          ['Maryland', 'MD'],
          ['Maine', 'ME'],
          ['Michigan', 'MI'],
          ['Minnesota', 'MN'],
          ['Missouri', 'MO'],
          ['Mississippi', 'MS'],
          ['Montana', 'MT'],
          ['North Carolina', 'NC'],
          ['North Dakota', 'ND'],
          ['Nebraska', 'NE'],
          ['New Hampshire', 'NH'],
          ['New Jersey', 'NJ'],
          ['New Mexico', 'NM'],
          ['Nevada', 'NV'],
          ['New York', 'NY'],
          ['Ohio', 'OH'],
          ['Oklahoma', 'OK'],
          ['Oregon', 'OR'],
          ['Pennsylvania', 'PA'],
          ['Rhode Island', 'RI'],
          ['South Carolina', 'SC'],
          ['South Dakota', 'SD'],
          ['Tennessee', 'TN'],
          ['Texas', 'TX'],
          ['Utah', 'UT'],
          ['Virginia', 'VA'],
          ['Vermont', 'VT'],
          ['Washington', 'WA'],
          ['Wisconsin', 'WI'],
          ['West Virginia', 'WV'],
          ['Wyoming', 'WY'],
          ['American Samoa', 'AS'],
          ['Guam', 'GU'],
          ['Puerto Rico', 'PR'],
          ['Virgin Islands', 'VI']
        ]
      )
    end
  end
end
